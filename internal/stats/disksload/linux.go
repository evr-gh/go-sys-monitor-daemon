//go:build linux

package disksload

import (
	"context"
	"strings"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/models"
	tools "github.com/evr-gh/go-sys-monitor-deamon/internal/tools"
)

func GetDisksLoad(ctx context.Context,
	exec func(ctx context.Context, command string, args []string) (string, error),
) (*models.DisksLoad, error) {
	disksLoad := make([]models.DiskLoad, 0, 4)

	res, err := exec(ctx, "iostat", []string{""})
	if err != nil {
		return nil, err
	}
	lines := strings.Split(res, "\n")
	for i := range lines {
		fields := strings.Fields(lines[i])
		if len(fields) < 8 || fields[0] == "Device" {
			continue
		}
		nDiskLoad := models.DiskLoad{
			Name: fields[0],

			Tps:      tools.ParseFloat(fields[1]),
			KpsRead:  tools.ParseFloat(fields[2]),
			KpsWrite: tools.ParseFloat(fields[3]),
		}
		disksLoad = append(disksLoad, nDiskLoad)
	}
	return &models.DisksLoad{DisksLoad: disksLoad}, nil
}
