package builder

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	SuperblockOffset  = 1024
	VolumeLabelOffset = 120
	MaxLabelLength    = 16
)

func SetExt4Label(imagePath string, label string) error {
	if len(label) > MaxLabelLength {
		return fmt.Errorf("label too long: %d bytes (max %d)", len(label), MaxLabelLength)
	}

	f, err := os.OpenFile(imagePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}
	defer f.Close()

	labelOffset := int64(SuperblockOffset + VolumeLabelOffset)
	if _, err := f.Seek(labelOffset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to label offset: %w", err)
	}

	labelBytes := make([]byte, MaxLabelLength)
	copy(labelBytes, []byte(label))

	if _, err := f.Write(labelBytes); err != nil {
		return fmt.Errorf("failed to write label: %w", err)
	}

	return f.Sync()
}

func VerifyExt4Magic(imagePath string) (bool, error) {
	f, err := os.Open(imagePath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	const MagicOffset = 0x438
	const Ext4Magic = 0xEF53

	if _, err := f.Seek(MagicOffset, io.SeekStart); err != nil {
		return false, err
	}

	var magic uint16
	if err := binary.Read(f, binary.LittleEndian, &magic); err != nil {
		return false, err
	}

	return magic == Ext4Magic, nil
}
