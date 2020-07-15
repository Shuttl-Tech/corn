package cmd

import (
	"context"
	"fmt"
	"github.com/Shuttl-Tech/corn/engine"
	"github.com/google/shlex"
	"github.com/hashicorp/cronexpr"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
)

// This command parses the arguments into Schedule.
// Parsing depends on the order of arguments as defined
// on command line. Consider for example:
//
//    corn manager --cmd first.command --sched "* * * * *" --cmd second.command --sched "1 * * * *"
//
// this command would be parsed into following schedule:
//
//    schedule {
//      "first.command": {RawExpr: "* * * * *", Expr: <*>},
//      "second.command": {RawExpr: "1 * * * *", Expr: <*>},
//    }
//
// It is important that flags are specified in `--cmd ... --sched ...` order
// for each individual cron or the parser will return an error.

type Expr struct {
	RawExpr string
	Expr    *cronexpr.Expression
}

type CmdFlags struct {
	schedule map[string]*Expr
}

func (f CmdFlags) Type() string {
	return "string"
}

func (f CmdFlags) String() string {
	return "<none>"
}

func (f CmdFlags) Set(val string) error {
	_, ok := f.schedule[val]
	if ok {
		return fmt.Errorf("duplicate command %q", val)
	}

	f.schedule[val] = nil
	return nil
}

type SchedFlags struct {
	schedule map[string]*Expr
}

func (f SchedFlags) Type() string {
	return "string"
}

func (f SchedFlags) String() string {
	return "<none>"
}

func (f SchedFlags) Set(val string) error {
	var recv string
	for cmd, exp := range f.schedule {
		if exp != nil {
			continue
		}

		if recv != "" {
			return fmt.Errorf(`extra command argument.
Possibly out of order '--sched' flag with value %q.
Order must be '--cmd ... --sched ...' for each cron.
Command %q and %q are awaiting --sched flag`, val, recv, cmd)
		}

		recv = cmd
	}

	if recv == "" {
		return fmt.Errorf("missing command argument. Possibly out of order `--sched` flag with value %q. Order must be `--cmd ... --sched ...` for each cron", val)
	}

	exp, err := cronexpr.Parse(val)
	if err != nil {
		return fmt.Errorf("invalid cron expression %q. %s", val, err)
	}

	f.schedule[recv] = &Expr{
		RawExpr: val,
		Expr:    exp,
	}

	return nil
}

var expressions = map[string]*Expr{}

var manager = &cobra.Command{
	Use:   "manager",
	Short: "Start cron manager",
	RunE:  runManager,
}

func init() {
	cmdF := &CmdFlags{schedule: expressions}
	schedF := &SchedFlags{schedule: expressions}

	manager.Flags().Var(cmdF, "cmd", "Command to execute")
	manager.Flags().Var(schedF, "sched", "Cron expression to compute the schedule of command")

	_ = manager.MarkFlagRequired("cmd")
	_ = manager.MarkFlagRequired("sched")

	baseCommand.AddCommand(manager)
}

func runManager(cmd *cobra.Command, args []string) error {
	var instructions []engine.Instruction
	for command, expr := range expressions {
		cmd, err := shlex.Split(command)
		if err != nil {
			return fmt.Errorf("malformed command %q. %s", command, err)
		}

		instructions = append(instructions, engine.Instruction{
			Expr:    expr.Expr,
			Command: cmd[0],
			Args:    cmd[1:],
		})
	}

	ctx := makeRunnerCtx()

	runner := engine.New(instructions)
	runner.Start(ctx)

	return nil
}

func makeRunnerCtx() context.Context {
	exit := make(chan os.Signal, 2)
	signal.Notify(exit, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-exit
		cancel()
	}()

	return ctx
}
