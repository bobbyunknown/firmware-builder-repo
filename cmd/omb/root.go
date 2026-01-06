package omb

import (
	cc "github.com/ivanpirog/coloredcobra"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "omb",
	Short: "Oh-my-builder - ARM firmware builder",
	Long:  "Centralized firmware builder for ARM devices with automated data synchronization",
}

func Execute() error {
	cc.Init(&cc.Config{
		RootCmd:  rootCmd,
		Headings: cc.HiCyan + cc.Bold + cc.Underline,
		Commands: cc.HiYellow + cc.Bold,
		Example:  cc.Italic,
		ExecName: cc.Bold,
		Flags:    cc.Bold,
	})

	return rootCmd.Execute()
}
