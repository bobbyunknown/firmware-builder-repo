package omb

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "omb",
	Short: "Oh-my-builder - ARM firmware builder",
	Long:  "Centralized firmware builder for ARM devices with automated data synchronization",
}

func Execute() error {
	return rootCmd.Execute()
}
