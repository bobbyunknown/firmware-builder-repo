package omb

import (
	"fmt"
	"log"

	"github.com/bobbyunknown/Oh-my-builder/pkg/download"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download firmware components",
	Long:  "Download kernels, rootfs, and other components from Oh-my-builder data branch",
}

var downloadKernelCmd = &cobra.Command{
	Use:   "kernel [version]",
	Short: "Download kernel files",
	Args:  cobra.ExactArgs(1),
	Run:   runDownloadKernel,
}

var downloadRootfsCmd = &cobra.Command{
	Use:   "rootfs [name]",
	Short: "Download rootfs file",
	Args:  cobra.ExactArgs(1),
	Run:   runDownloadRootfs,
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.AddCommand(downloadKernelCmd)
	downloadCmd.AddCommand(downloadRootfsCmd)
}

func runDownloadKernel(cmd *cobra.Command, args []string) {
	version := args[0]

	dm, err := download.NewManager()
	if err != nil {
		log.Fatalf("Failed to create download manager: %v", err)
	}

	if err := dm.DownloadKernel(version); err != nil {
		log.Fatalf("Failed to download kernel: %v", err)
	}

	fmt.Printf("\n✓ Kernel %s downloaded to %s\n", version, dm.GetKernelPath(version))
}

func runDownloadRootfs(cmd *cobra.Command, args []string) {
	name := args[0]

	dm, err := download.NewManager()
	if err != nil {
		log.Fatalf("Failed to create download manager: %v", err)
	}

	if err := dm.DownloadRootfs(name); err != nil {
		log.Fatalf("Failed to download rootfs: %v", err)
	}

	fmt.Printf("\n✓ Rootfs %s downloaded to %s\n", name, dm.GetRootfsPath(name))
}
