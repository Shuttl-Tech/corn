package cmd

import (
	"github.com/spf13/cobra"
	"log"
)

var Version string

var baseCommand = &cobra.Command{
	Use:   "Corn",
	Short: "Corn is an infrastructure abstraction layer to manage cron jobs",
	PersistentPreRun: func(*cobra.Command, []string) {
		log.Printf("Corn version %s", Version)
	},
}

func Execute() error {
	return baseCommand.Execute()
}
