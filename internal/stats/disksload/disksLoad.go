package disksload

import (
	"context"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/models"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/tools"
)

func GetStats(ctx context.Context) (*models.DisksLoad, error) {
	diskLoad, err := GetDisksLoad(ctx, tools.ExecCommand)
	return diskLoad, err
}
