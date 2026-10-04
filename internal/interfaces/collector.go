package interfaces

import (
	"context"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/server/rpc/grpcapi"
)

type Collector interface {
	Run(ctx context.Context) error
	Stop()

	RegisterRequest(averagingPeriod int64, statTyes []grpcapi.StatType) error
	OnRequestCancel(averagingPeriod int64, statTyes []grpcapi.StatType)

	PrepareResponse(averagingPeriod int64, statTypes []grpcapi.StatType) *grpcapi.StatsResponse
}
