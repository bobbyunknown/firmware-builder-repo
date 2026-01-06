package omb

import (
	"fmt"
	"log"
	"os"

	"github.com/bobbyunknown/Oh-my-builder/pkg/builder"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build firmware image",
	Long:  "Build firmware image using profile file, flags, or interactive mode",
	Run:   runBuild,
}

var (
	profileFile string
	deviceFlag  string
	kernelFlag  string
	rootfsFlag  string
	sizeFlag    int
	outputFlag  string
	patchFlag   string
)

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().StringVarP(&profileFile, "profile", "p", "", "Build profile file")
	buildCmd.Flags().StringVarP(&deviceFlag, "device", "d", "", "Device name")
	buildCmd.Flags().StringVarP(&kernelFlag, "kernel", "k", "", "Kernel version")
	buildCmd.Flags().StringVarP(&rootfsFlag, "rootfs", "r", "", "Rootfs file")
	buildCmd.Flags().IntVarP(&sizeFlag, "size", "s", 1024, "Image size in MB")
	buildCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output file path")
	buildCmd.Flags().StringVar(&patchFlag, "patch", "", "Patch archive name")
}

func runBuild(cmd *cobra.Command, args []string) {
	var config builder.BuildConfig

	if profileFile != "" {
		if err := loadProfile(profileFile, &config); err != nil {
			log.Fatalf("Failed to load profile: %v", err)
		}
	} else if deviceFlag != "" && kernelFlag != "" && rootfsFlag != "" {
		config = builder.BuildConfig{
			Device: deviceFlag,
			Kernel: kernelFlag,
			Rootfs: rootfsFlag,
			Size:   sizeFlag,
			Output: outputFlag,
		}
	} else {
		log.Fatal("Either --profile or --device/--kernel/--rootfs flags are required")
	}

	if config.Output == "" {
		config.Output = fmt.Sprintf("out/%s.img", config.Device)
	}

	b, err := builder.NewBuilder(config, ".cache")
	if err != nil {
		log.Fatalf("Failed to create builder: %v", err)
	}
	defer b.Cleanup()

	if err := b.Build(); err != nil {
		log.Fatalf("Build failed: %v", err)
	}
}

func loadProfile(path string, config *builder.BuildConfig) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, config)
}
