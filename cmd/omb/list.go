package omb

import (
	"fmt"
	"log"

	"github.com/bobbyunknown/Oh-my-builder/pkg/index"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available components",
	Long:  "List available kernels, rootfs, or devices from cached indexes",
}

var listKernelsCmd = &cobra.Command{
	Use:   "kernels",
	Short: "List available kernels",
	Run:   runListKernels,
}

var listDevicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List available devices",
	Run:   runListDevices,
}

var vendorFilter string

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(listKernelsCmd)
	listCmd.AddCommand(listDevicesCmd)

	listDevicesCmd.Flags().StringVar(&vendorFilter, "vendor", "", "Filter devices by vendor (amlogic, rockchip, allwinner)")
}

func runListKernels(cmd *cobra.Command, args []string) {
	// TODO: Implement kernel listing
	fmt.Println("Kernel listing not yet implemented")
	fmt.Println("Run './omb repo update' first to fetch kernel list")
}

func runListDevices(cmd *cobra.Command, args []string) {
	registry, err := index.LoadDevices("configs/devices.yaml")
	if err != nil {
		log.Fatalf("Failed to load devices: %v\nRun './omb repo update' to fetch device list", err)
	}

	var devices []index.Device
	if vendorFilter != "" {
		devices = registry.FindByVendor(vendorFilter)
		fmt.Printf("Available Devices (%s):\n", vendorFilter)
	} else {
		devices = registry.ListAll()
		fmt.Println("Available Devices:")
	}

	if len(devices) == 0 {
		fmt.Println("No devices found")
		return
	}

	for _, device := range devices {
		fmt.Printf("  %-30s (Vendor: %s)\n", device.Name, device.Vendor)
	}

	fmt.Printf("\nTotal: %d devices\n", len(devices))
}
