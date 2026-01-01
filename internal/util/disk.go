package util

import (
	"os"
	"regexp"
)

const (
	DiskByIdPath = "/dev/disk/by-id/"
)

type SmartCtlData struct {
}

type DiskInfo struct {
	Name         string
	Path         string
	SmartCtlData SmartCtlData
}

func GetDisks() ([]DiskInfo, error) {
	result := []DiskInfo{}

	files, err := os.ReadDir(DiskByIdPath)
	if err != nil {
		return result, err
	}

	regex := regexp.MustCompile(`^(ata|^nvme|^scsi|^wwn)-[^-]*$`)

	for _, f := range files {
		if f.Type() != os.ModeSymlink {
			continue
		}

		// resolve symlink
		entryPath := DiskByIdPath + f.Name()
		linkTarget, err := os.Readlink(entryPath)
		if err != nil {
			continue
		}

		if !regex.MatchString(f.Name()) {
			continue
		}

		// if link target is already in the result list, skip it
		skip := false
		for _, d := range result {
			if d.Path == linkTarget {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		disk := DiskInfo{
			Name: f.Name(),
			Path: linkTarget,
		}
		result = append(result, disk)
	}

	return result, nil
}
