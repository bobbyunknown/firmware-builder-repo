# Build Profile Format

## Overview

Build profiles are YAML files that define firmware build configurations. They allow reproducible builds and easy sharing of build configurations.

## Profile Structure

```yaml
device: <device-name>
kernel: <kernel-version>
rootfs: <rootfs-filename>
size: <image-size-mb>
output: <output-path>
```

## Fields

### device (required)

Device name from the devices index.

**Example:**
```yaml
device: h616-x96-mate
```

**Available devices:**
```bash
./omb list devices
```

### kernel (required)

Kernel version from the kernels index.

**Example:**
```yaml
kernel: 6.1.123
```

**Available kernels:**
```bash
./omb list kernels
```

### rootfs (required)

Rootfs filename from the rootfs index.

**Example:**
```yaml
rootfs: openwrt-23.05.5-vanila-armsr-armv8-generic-ext4-rootfs.img.gz
```

**Available rootfs:**
```bash
./omb repo update
cat configs/rootfs.yaml
```

### size (required)

Image size in megabytes (MB).

**Recommended sizes:**
- Minimal: 512 MB
- Standard: 1024 MB (1 GB)
- Extended: 2048 MB (2 GB)
- Large: 4096 MB (4 GB)

**Example:**
```yaml
size: 1024
```

### output (optional)

Output file path. If not specified, defaults to `out/<device-name>.img`.

**Example:**
```yaml
output: out/custom-name.img
```

## Example Profiles

### Allwinner H616 - OpenWrt

```yaml
device: h616-x96-mate
kernel: 6.1.123
rootfs: openwrt-23.05.5-vanila-armsr-armv8-generic-ext4-rootfs.img.gz
size: 1024
output: out/h616-openwrt.img
```

### Rockchip RK3588 - OpenWrt

```yaml
device: rk3588-orangepi-5-plus
kernel: 6.1.123
rootfs: openwrt-23.05.5-vanila-armsr-armv8-generic-ext4-rootfs.img.gz
size: 2048
output: out/rk3588-openwrt.img
```

### Amlogic S905x3 - OpenWrt

```yaml
device: s905x3
kernel: 5.4.279
rootfs: openwrt-23.05.5-vanila-armsr-armv8-generic-ext4-rootfs.img.gz
size: 1024
output: out/s905x3-openwrt.img
```

### Allwinner H618 - Custom Rootfs

```yaml
device: h618-orangepi-zero3
kernel: 6.6.6-AW64-DBAI
rootfs: OpenWrt-23.05.5-ALKHANET-armsr-armv8-generic-rootfs.tar.gz
size: 2048
output: out/h618-alkhanet.img
```

## Profile Validation

Before building, the tool validates:

1. Device exists in devices index
2. Kernel exists in cache (auto-downloads if missing)
3. Rootfs exists in cache (auto-downloads if missing)
4. Output directory is writable
5. Sufficient disk space available

## Profile Management

### Creating Profiles

Store profiles in `profiles/` directory:

```
profiles/
├── production/
│   ├── h616-stable.yaml
│   └── rk3588-stable.yaml
├── testing/
│   ├── h616-beta.yaml
│   └── rk3588-beta.yaml
└── examples/
    ├── h616-openwrt.yaml
    ├── rk3588-openwrt.yaml
    └── s905x3-openwrt.yaml
```

### Using Profiles

```bash
./omb build -p profiles/production/h616-stable.yaml
```

### Sharing Profiles

Profiles are plain YAML files and can be:
- Committed to version control
- Shared via email or chat
- Published in documentation

## Best Practices

1. **Use descriptive names**: `h616-openwrt-stable.yaml` instead of `build1.yaml`
2. **Organize by purpose**: production/, testing/, examples/
3. **Document custom settings**: Add comments in YAML
4. **Version control**: Track profile changes in git
5. **Test before production**: Use testing profiles first

## Advanced Usage

### Comments

Add comments to document your configuration:

```yaml
device: h616-x96-mate
kernel: 6.1.123
rootfs: openwrt-23.05.5-vanila.img.gz
size: 1024
output: out/h616-openwrt.img
```

### Multiple Profiles

Create multiple profiles for different use cases:

```bash
./omb build -p profiles/minimal.yaml
./omb build -p profiles/full-featured.yaml
./omb build -p profiles/testing.yaml
```
