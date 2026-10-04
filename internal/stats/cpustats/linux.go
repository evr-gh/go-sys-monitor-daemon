//go:build linux

package cpustats

import (
	"context"
	"strings"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/models"
	tools "github.com/evr-gh/go-sys-monitor-deamon/internal/tools"
)

const (
	userPos   = 14
	systemPos = 16
	idlePos   = 19
)

func GetCPUStat(ctx context.Context,
	exec func(ctx context.Context, command string, args []string) (string, error),
) (*models.CPUStat, error) {
	res, err := exec(ctx, "iostat", []string{"-c"})
	if err != nil {
		return nil, err
	}

	fields := strings.Fields(res)

	user := tools.ParseFloat(fields[userPos])
	system := tools.ParseFloat(fields[systemPos])
	idle := tools.ParseFloat(fields[idlePos])

	return &models.CPUStat{
		UserMode:   user,
		SystemMode: system,
		Idle:       idle,
	}, nil
}
