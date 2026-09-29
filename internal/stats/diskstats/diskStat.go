package diskstats

import (
	"context"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/models"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/tools"
)

func GetStats(ctx context.Context) (*models.DiskStats, error) {
	diskStat, err := GetDiskStats(ctx, tools.ExecCommand)
	return diskStat, err
}
