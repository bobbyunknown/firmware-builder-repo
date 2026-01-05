package omb

import (
	"fmt"
	"log"

	"github.com/bobbyunknown/Oh-my-builder/pkg/repo"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify data repository structure",
	Long:  "Check if all required components exist in the data repository",
	Run:   runVerify,
}

func init() {
	repoCmd.AddCommand(verifyCmd)
}

func runVerify(cmd *cobra.Command, args []string) {
	checker := repo.NewDataChecker("bobbyunknown", "Oh-my-builder", "data")

	if err := checker.VerifyDataStructure(); err != nil {
		log.Fatalf("Verification failed: %v", err)
	}

	fmt.Println("\n✅ All components verified successfully!")
}
