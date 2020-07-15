package engine

import (
	"context"
	"github.com/hashicorp/cronexpr"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"time"
)

type Instruction struct {
	Expr    *cronexpr.Expression
	Command string
	Args    []string
}

type Engine struct {
	instructions []Instruction
	logger       *log.Logger
}

// TODO: consider exporting metrics to datadog.
//    Crons can emit very high cardinality metrics because
//    of tagging. We'll revisit stats after judging
//    the impact and necessity from logs
func New(inst []Instruction) *Engine {
	logger := log.New()
	logger.SetLevel(log.InfoLevel)
	logger.SetFormatter(&log.JSONFormatter{})

	return &Engine{
		instructions: inst,
		logger:       logger,
	}
}

func (engine *Engine) Start(ctx context.Context) {
	engine.logger.WithField("instructions", len(engine.instructions)).Info("starting corn engine")
	for _, inst := range engine.instructions {
		go engine.handleInstruction(ctx, inst)
	}

	<-ctx.Done()
	engine.logger.Info("corn context is done")
}

func (engine *Engine) handleInstruction(ctx context.Context, inst Instruction) {
	nextExec := inst.Expr.Next(time.Now())
	if nextExec.IsZero() {
		engine.logger.
			WithField("command", inst.Command).
			WithField("args", inst.Args).
			Errorf("cron expression evaluated to zero time. shutting down the executor thread")
		return
	}

	threadLog := engine.logger.
		WithField("command", inst.Command).
		WithField("args", inst.Args)

	next := nextExec.Sub(time.Now())
	ticker := time.NewTimer(next)

	threadLog.WithField("next_execution", nextExec).
		WithField("next_exec_remaining", next.String()).
		Info("starting cron executor thread")

	for {
		select {
		case <-ctx.Done():
			threadLog.Info("executor context is done")
		case <-ticker.C:
			nextExec = inst.Expr.Next(time.Now())
			next = nextExec.Sub(time.Now())
			ticker.Reset(next)

			// TODO: make deadline tolerance configurable or
			//    implement different strategies for different
			//    periods. 80% of the schedule may be too low
			//    for crons with short periodicity and too high
			//    for crons with long periodicity.
			deadline := (next * 80) / 100

			threadLog.WithField("deadline", deadline).Info("starting cron execution")
			cmdCtx, _ := context.WithTimeout(ctx, deadline)

			cmd := exec.CommandContext(cmdCtx, inst.Command, inst.Args...)
			cmd.Stdout = threadLog.Writer()
			cmd.Stderr = threadLog.Writer()

			err := cmd.Run()
			if err != nil {
				threadLog.WithError(err).Error("cron execution failed")
			}

			threadLog.WithField("next_execution", nextExec).
				WithField("next_exec_remaining", next.String()).
				Info("scheduling cron for next execution")
		}
	}
}
