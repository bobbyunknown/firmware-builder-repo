# Build Guide

## Overview

Oh-my-builder supports three ways to build firmware images:
1. **Profile File** - Recommended for reproducible builds
2. **CLI Flags** - Quick testing and one-off builds
3. **Interactive Mode** - User-friendly for beginners (coming soon)

## Build Modes

### 1. Profile File Mode

Create a YAML profile file with your build configuration:

```yaml
device: h616-x96-mate
kernel: 6.1.123
rootfs: openwrt-23.05.5-vanila-armsr-armv8-generic-ext4-rootfs.img.gz
size: 1024
output: out/h616-openwrt.img
patch: startup.tar.xz
```

Build using the profile:

```bash
./omb build -p profiles/my-build.yaml
```

### 2. CLI Flags Mode

Build directly with command-line flags:

```bash
./omb build \
  --device h616-x96-mate \
  --kernel 6.1.123 \
  --rootfs openwrt-23.05.5-vanila.img.gz \
  --size 1024 \
  --output out/custom.img \
  --patch startup.tar.xz
```

### 3. Interactive Mode (Coming Soon)

Simply run:

```bash
./omb build
```

The tool will prompt you for each option.

## Build Process

The build process follows these steps:

1. **Validation** - Check if device, kernel, and rootfs exist (auto-downloads if missing)
2. **Create Image** - Create disk image with partitions
3. **Install Kernel** - Extract and install kernel files
4. **Install Rootfs** - Extract and install root filesystem
5. **Write Bootloader** - Write vendor-specific bootloader
6. **Finalize** - Compress and save final image

## Example Workflows

### Build for Allwinner H616

```bash
./omb repo update
# No need to manually download kernel/rootfs, builder will do it automatically!
./omb build -p profiles/examples/h616-openwrt.yaml
```

### Build for Rockchip RK3588

```bash
./omb build \
  -d rk3588-orangepi-5-plus \
  -k 6.1.123 \
  -r openwrt-23.05.5.img.gz \
  -s 2048 \
  -o out/rk3588.img \
  --patch startup.tar.xz
```

### Build for Amlogic S905x3

```bash
./omb build -p profiles/examples/s905x3-openwrt.yaml
```

## Output

The build process creates:

```
out/
└── device-name.img       # Final firmware image
```

You can then flash this image to SD card or eMMC:

```bash
sudo dd if=out/h616-openwrt.img of=/dev/sdX bs=4M status=progress
```

## Troubleshooting

### Auto-Download Failed

```
Error: failed to auto-download kernel: ...
```

**Solution:** Check your internet connection or try downloading manually:
```bash
./omb download kernel 6.1.123
```

### Insufficient disk space

```
Error: not enough space to create image
```

**Solution:** Free up disk space or reduce image size in profile.

## Advanced Options

### Custom Image Size

Adjust the `size` parameter (in MB):

```yaml
size: 2048  # 2GB image
```

### Custom Output Path

Specify custom output location:

```yaml
output: /path/to/custom/location.img
```

## Next Steps

- See [PROFILES.md](PROFILES.md) for profile file format details
- See [BOOTLOADERS.md](BOOTLOADERS.md) for bootloader information
- See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues
