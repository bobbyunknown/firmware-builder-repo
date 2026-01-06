package omb

import (
	"fmt"
	"log"

	"github.com/bobbyunknown/Oh-my-builder/pkg/download"
	"github.com/bobbyunknown/Oh-my-builder/pkg/repo"
	"github.com/spf13/cobra"
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage repository indexes",
	Long:  "Update and manage local repository indexes from Oh-my-builder data branch",
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update repository indexes",
	Long:  "Fetch latest metadata from Oh-my-builder/data and update local YAML indexes",
	Run:   runUpdate,
}

func init() {
	rootCmd.AddCommand(repoCmd)
	repoCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) {
	fmt.Println("🔄 Updating repository indexes...")
	fmt.Println()

	repoCfg, err := repo.LoadDataRepo("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	owner, name, err := repo.ParseRepoURL(repoCfg.URL)
	if err != nil {
		log.Fatalf("Failed to parse repo url: %v", err)
	}
	indexer := repo.NewIndexer(owner, name, repoCfg.Branch, repoCfg.Components)

	fmt.Print("   Fetching kernels...     ")
	kernels, err := indexer.FetchKernelIndex()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✓ Found %d kernels\n", len(kernels.Kernels))
	}

	fmt.Print("   Fetching rootfs...      ")
	rootfs, err := indexer.FetchRootfsIndex()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✓ Found %d rootfs\n", len(rootfs.Rootfs))
	}

	fmt.Print("   Fetching devices...     ")
	devices, err := indexer.FetchDeviceIndex()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✓ Found %d devices\n", len(devices.Devices))
	}

	fmt.Print("   Fetching patch...       ")
	patches, err := indexer.FetchPatchIndex()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✓ Found %d patches\n", len(patches.Patches))
	}

	fmt.Println()
	fmt.Println("📝 Saving indexes...")

	if kernels != nil {
		fmt.Print("   kernels.yaml            ")
		if err := repo.SaveIndex("configs/kernels.yaml", kernels); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Println("✓ Saved")
		}
	}

	if rootfs != nil {
		fmt.Print("   rootfs.yaml             ")
		if err := repo.SaveIndex("configs/rootfs.yaml", rootfs); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Println("✓ Saved")
		}
	}

	if devices != nil {
		fmt.Print("   devices.yaml            ")
		if err := repo.SaveIndex("configs/devices.yaml", devices); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Println("✓ Saved")
		}
	}

	if patches != nil {
		fmt.Print("   patch.yaml              ")
		if err := repo.SaveIndex("configs/patch.yaml", patches); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Println("✓ Saved")
		}
	}

	fmt.Println()
	fmt.Println("✨ Repository indexes updated successfully!")

	dm, err := download.NewManager()
	if err == nil {
		if err := dm.ValidateCache(); err != nil {
			fmt.Printf("Warning: cache validation failed: %v\n", err)
		}
	}
}
