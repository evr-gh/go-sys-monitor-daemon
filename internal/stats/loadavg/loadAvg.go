package loadavg

import (
	"context"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/models"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/tools"
)

func GetStats(ctx context.Context) (*models.LoadAverage, error) {
	loadAvg, err := GetLoadAvg(ctx, tools.Exec)

	return loadAvg, err
}
