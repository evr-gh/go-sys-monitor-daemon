package diskstats

import (
	"context"
	"fmt"
	"strings"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/models"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/tools"
)

func GetDiskStats(ctx context.Context,
	exec func(ctx context.Context, command string, args []string) (string, error),
) (*models.DiskStats, error) {
	mbOut, err := exec(ctx, "df", []string{
		"-T", "-m", "--exclude-type=tmpfs",
		"--exclude-type=devtmpfs", "--exclude-type=udev",
	})
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(mbOut), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("не полученны данные об использовании дисков в Mб")
	}
	devices := lines[1:]
	output := make([]models.DiskStat, 0, len(devices))

	inOut, err := exec(ctx, "df", []string{
		"-T", "-i", "--exclude-type=tmpfs",
		"--exclude-type=devtmpfs", "--exclude-type=udev",
	})
	if err != nil {
		return nil, err
	}

	inLines := strings.Split(strings.TrimSpace(inOut), "\n")
	if len(inLines) != len(lines) {
		return nil, fmt.Errorf("информация об использовании I-нод не соответствует иинформации об использовании дисков в Mб")
	}

	for i, disk := range devices {
		diskArr := strings.Fields(disk)
		diskInodeArr := strings.Fields(inLines[i+1])

		if len(diskArr) < 6 || (len(diskInodeArr) < 6) {
			continue
		}
		if diskArr[0] != diskInodeArr[0] {
			continue
		}

		diskStat := models.DiskStat{
			Name:       diskArr[0],
			MBUsagePct: tools.ParseUint32(strings.ReplaceAll(diskArr[5], "%", "")),
			InUsagePct: tools.ParseUint32(strings.ReplaceAll(diskInodeArr[5], "%", "")),
		}

		output = append(output, diskStat)
	}

	return &models.DiskStats{DiskStats: output}, nil
}
