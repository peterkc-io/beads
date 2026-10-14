package main

import (
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const programName = "bdx"

var bdxHelpOnce sync.Once

// rewriteBdxHelp changes only Cobra's help metadata, after all commands register.
func rewriteBdxHelp() {
	bdxHelpOnce.Do(func() {
		var walk func(*cobra.Command)
		walk = func(cmd *cobra.Command) {
			cmd.Use = strings.ReplaceAll(cmd.Use, "bd ", programName+" ")
			cmd.Short = strings.ReplaceAll(cmd.Short, "bd ", programName+" ")
			cmd.Long = strings.ReplaceAll(cmd.Long, "bd ", programName+" ")
			cmd.Example = strings.ReplaceAll(cmd.Example, "bd ", programName+" ")
			rewriteFlag := func(flag *pflag.Flag) {
				flag.Usage = strings.ReplaceAll(flag.Usage, "bd ", programName+" ")
			}
			cmd.LocalFlags().VisitAll(rewriteFlag)
			cmd.PersistentFlags().VisitAll(rewriteFlag)
			for _, child := range cmd.Commands() {
				walk(child)
			}
		}
		walk(rootCmd)
	})
}
