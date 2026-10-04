package cpustats

import (
	"context"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/models"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/tools"
)

func GetStats(ctx context.Context) (*models.CPUStat, error) {
	cpuInfo, err := GetCPUStat(ctx, tools.ExecCommand)
	return cpuInfo, err
}
