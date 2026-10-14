package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestBdxHelpCommandTree(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()
	var save func(*cobra.Command)
	save = func(cmd *cobra.Command) {
		use, short, long, example := cmd.Use, cmd.Short, cmd.Long, cmd.Example
		t.Cleanup(func() {
			cmd.Use, cmd.Short, cmd.Long, cmd.Example = use, short, long, example
		})
		saveFlag := func(flag *pflag.Flag) {
			usage := flag.Usage
			t.Cleanup(func() { flag.Usage = usage })
		}
		cmd.LocalFlags().VisitAll(saveFlag)
		cmd.PersistentFlags().VisitAll(saveFlag)
		for _, child := range cmd.Commands() {
			save(child)
		}
	}
	save(rootCmd)
	var walk func(*cobra.Command)
	walk = func(cmd *cobra.Command) {
		t.Run(cmd.CommandPath(), func(t *testing.T) {
			out := captureStdout(t, func() error {
				cmd.HelpFunc()(cmd, []string{"--help"})
				return nil
			})
			if strings.Contains(out, "bd ") {
				t.Errorf("rendered --help still contains a bd command reference")
			}
			if !strings.Contains(out, "bdx") {
				t.Errorf("rendered --help does not name bdx")
			}
		})
		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(rootCmd)
}
