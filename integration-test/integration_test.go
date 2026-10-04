package integration

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/collector"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/config"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/interfaces"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/logger"
	rpcServer "github.com/evr-gh/go-sys-monitor-deamon/internal/server/rpc"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/server/rpc/grpcapi"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func testCase(t *testing.T,
	cmdConfig config.DaemonConfig,
	serverTimeout int64,
	clientTimeout int64,
	statTypes []grpcapi.StatType,
	interval int64,
	averagingPeriod int64,
) {
	t.Helper()

	logg := logger.New(cmdConfig.Logger.Level, os.Stdout)

	metricsCollector := collector.NewMetricsCollector(logg, cmdConfig.Stats)

	rpcServer := rpcServer.NewRPCServer(logg, metricsCollector)

	srvCtx, srvCancel := context.WithTimeout(context.Background(), time.Duration(float64(serverTimeout)+2.001)*time.Second)
	defer srvCancel()

	wg := sync.WaitGroup{}

	var once sync.Once

	wg.Go(func() {
		<-srvCtx.Done()
		rpcServer.GracefulStop()
	})

	var err1 error
	wg.Go(func() {
		if err1 = rpcServer.Start(srvCtx, fmt.Sprintf("%s:%d", cmdConfig.RPC.Host, cmdConfig.RPC.Port)); err1 != nil {
			once.Do(srvCancel)
		}
	})

	var err2 error
	wg.Go(func() {
		logg.Info("Начало работы сервиса мониторинга")
		err2 = metricsCollector.Run(srvCtx)
		logg.Info("Завершение работы сервиса мониторинга")
	})

	time.Sleep(2 * time.Second)

	cltCtx, CltCancel := context.WithTimeout(context.Background(), time.Duration(clientTimeout)*time.Second)
	defer CltCancel()

	addr := fmt.Sprintf("%s:%d", cmdConfig.RPC.Host, cmdConfig.RPC.Port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoErrorf(t, err, "Не удалось подключиться к %s: %v", addr, err)

	defer conn.Close()

	client := grpcapi.NewStatsServiceClient(conn)

	req := &grpcapi.StatsRequest{
		Interval:        interval,
		AveragingPeriod: averagingPeriod,
		StatTypes:       statTypes,
	}

	stream, err := client.GetStats(cltCtx, req)
	require.NoErrorf(t, err, "Не удалось получить статистику: %v", err)

	logg.Info("Послан запрос статистики")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	respCount := 0
	for {
		select {
		case <-cltCtx.Done():
			log.Printf("Прервана работа клиента")
			goto END
		case <-ticker.C:
			resp, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					logg.Info("Сервер завершил поток")
					goto END
				}
				st, ok := status.FromError(err)
				require.Equalf(t, true, ok, "Получена не-gRPC ошибка: %v", err)

				if st.Code() != codes.DeadlineExceeded {
					require.NoErrorf(t, err, "Ошибка сервера: code=%s, message=%q",
						st.Code(),
						st.Message())
				}
				goto ERR_END
			} else {
				respCount++
				logg.Info("Получена статистика: %+v\n", resp)
				for _, st := range statTypes {
					switch st {
					case grpcapi.StatType_LOAD_AVERAGE:
						require.NotNil(t, resp.GetLoadAverage())
						require.True(t, resp.GetLoadAverage().Load1Min >= 0)
						require.True(t, resp.GetLoadAverage().Load5Min >= 0)
						require.True(t, resp.GetLoadAverage().Load15Min >= 0)
						logg.Info("Получена статистика LoadAverage: %+v\n", resp.GetLoadAverage())
					case grpcapi.StatType_CPU_STATS:
						require.NotNil(t, resp.GetCpuStats())
						require.True(t, resp.GetCpuStats().Idle >= 0)
						require.True(t, resp.GetCpuStats().SystemMode >= 0)
						require.True(t, resp.GetCpuStats().UserMode >= 0)
						logg.Info("Получена статистика CpuStats: %+v\n", resp.GetCpuStats())
					case grpcapi.StatType_DISKS_LOAD:
						require.NotNil(t, resp.GetDisksLoad())
						require.NotEmpty(t, resp.GetDisksLoad().GetDiskLoad())
						logg.Info("Получена статистика DisksLoad: %+v\n", resp.GetDisksLoad())
					case grpcapi.StatType_DISKS_STATS:
						require.NotNil(t, resp.GetDisksStats())
						require.NotEmpty(t, resp.GetDisksStats().GetDiskStats())
						logg.Info("Получена статистика DisksStats: %+v\n", resp.GetDisksStats())
					case grpcapi.StatType_NETWORK_TOP_TALKERS:
					case grpcapi.StatType_NETWORK_CONN_STATS:
					}
				}
			}
		}
	}

END:
	require.Equal(t, (clientTimeout-averagingPeriod)/interval, respCount)
ERR_END:
	wg.Wait()

	if err1 != nil {
		require.NoErrorf(t, err1, "Не удалось запустить RPC сервер: %v", err1.Error())
	}
	if err2 != nil {
		require.NoErrorf(t, err2, "Ошибка в работе сервиса мониторинга: %v", err2.Error())
	}
}

func testCase2(t *testing.T,
	cmdConfig config.DaemonConfig,
	serverTimeout int64,
	clientTimeout int64,
	statTypes []grpcapi.StatType,
	interval int64,
	averagingPeriod int64,
) {
	t.Helper()

	logg := logger.New(cmdConfig.Logger.Level, os.Stdout)

	metricsCollector := collector.NewMetricsCollector(logg, cmdConfig.Stats)

	rpcServer := rpcServer.NewRPCServer(logg, metricsCollector)

	srvCtx, srvCancel := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer srvCancel()

	wg := sync.WaitGroup{}

	var once sync.Once

	wg.Go(func() {
		<-srvCtx.Done()
		rpcServer.GracefulStop()
	})

	var err1 error
	wg.Go(func() {
		if err1 = rpcServer.Start(srvCtx, fmt.Sprintf("%s:%d", cmdConfig.RPC.Host, cmdConfig.RPC.Port)); err1 != nil {
			once.Do(srvCancel)
		}
	})

	var err2 error
	wg.Go(func() {
		logg.Info("Начало работы сервиса мониторинга")
		err2 = metricsCollector.Run(srvCtx)
		logg.Info("Завершение работы сервиса мониторинга")
	})

	wg.Go(func() {
		time.Sleep(time.Duration(float64(serverTimeout)+2.0001) * time.Second)
		pid, _, _ := syscall.Syscall(syscall.SYS_GETPID, 0, 0, 0)
		process, _ := os.FindProcess(int(pid))
		process.Signal(syscall.SIGHUP)
	})

	time.Sleep(2 * time.Second)

	cltCtx, CltCancel := context.WithTimeout(context.Background(), time.Duration(clientTimeout)*time.Second)
	defer CltCancel()

	addr := fmt.Sprintf("%s:%d", cmdConfig.RPC.Host, cmdConfig.RPC.Port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoErrorf(t, err, "Не удалось подключиться к %s: %v", addr, err)

	defer conn.Close()

	client := grpcapi.NewStatsServiceClient(conn)

	req := &grpcapi.StatsRequest{
		Interval:        interval,
		AveragingPeriod: averagingPeriod,
		StatTypes:       statTypes,
	}

	stream, err := client.GetStats(cltCtx, req)
	require.NoErrorf(t, err, "Не удалось получить статистику: %v", err)

	logg.Info("Послан запрос статистики")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	respCount := 0
	for {
		select {
		case <-cltCtx.Done():
			log.Printf("Прервана работа клиента")
			goto END
		case <-ticker.C:

			resp, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					logg.Info("Сервер завершил поток")
					goto END
				}
				st, ok := status.FromError(err)
				require.Equalf(t, true, ok, "Получена не-gRPC ошибка: %v", err)

				if st.Code() != codes.DeadlineExceeded {
					require.NoErrorf(t, err, "Ошибка сервера: code=%s, message=%q",
						st.Code(),
						st.Message())
				}
				goto ERR_END
			} else {
				respCount++
				logg.Info("Получена статистика: %+v\n", resp)
				for _, st := range statTypes {
					switch st {
					case grpcapi.StatType_LOAD_AVERAGE:
						require.NotNil(t, resp.GetLoadAverage())
						require.True(t, resp.GetLoadAverage().Load1Min >= 0)
						require.True(t, resp.GetLoadAverage().Load5Min >= 0)
						require.True(t, resp.GetLoadAverage().Load15Min >= 0)
						logg.Info("Получена статистика LoadAverage: %+v\n", resp.GetLoadAverage())
					case grpcapi.StatType_CPU_STATS:
						require.NotNil(t, resp.GetCpuStats())
						require.True(t, resp.GetCpuStats().Idle >= 0)
						require.True(t, resp.GetCpuStats().SystemMode >= 0)
						require.True(t, resp.GetCpuStats().UserMode >= 0)
						logg.Info("Получена статистика CpuStats: %+v\n", resp.GetCpuStats())
					case grpcapi.StatType_DISKS_LOAD:
						require.NotNil(t, resp.GetDisksLoad())
						require.NotEmpty(t, resp.GetDisksLoad().GetDiskLoad())
						logg.Info("Получена статистика DisksLoad: %+v\n", resp.GetDisksLoad())
					case grpcapi.StatType_DISKS_STATS:
						require.NotNil(t, resp.GetDisksStats())
						require.NotEmpty(t, resp.GetDisksStats().GetDiskStats())
						logg.Info("Получена статистика DisksStats: %+v\n", resp.GetDisksStats())
					case grpcapi.StatType_NETWORK_TOP_TALKERS:
					case grpcapi.StatType_NETWORK_CONN_STATS:
					}
				}
			}
		}
	}
END:

	require.Equal(t, (clientTimeout-averagingPeriod)/interval, respCount)
ERR_END:
	wg.Wait()

	if err1 != nil {
		require.NoErrorf(t, err1, "Не удалось запустить RPC сервер: %v", err1.Error())
	}
	if err2 != nil {
		require.NoErrorf(t, err2, "Ошибка в работе сервиса мониторинга: %v", err2.Error())
	}
}

func testCase3(t *testing.T,
	cmdConfig config.DaemonConfig,
	serverTimeout int64,
	clientTimeout int64,
	statTypes []grpcapi.StatType,
	interval int64,
	averagingPeriod int64,
	code codes.Code,
) {
	t.Helper()

	logg := logger.New(cmdConfig.Logger.Level, os.Stdout)

	metricsCollector := collector.NewMetricsCollector(logg, cmdConfig.Stats)

	rpcServer := rpcServer.NewRPCServer(logg, metricsCollector)

	srvCtx, srvCancel := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer srvCancel()

	wg := sync.WaitGroup{}

	var once sync.Once

	wg.Go(func() {
		<-srvCtx.Done()
		rpcServer.GracefulStop()
	})

	var err1 error
	wg.Go(func() {
		if err1 = rpcServer.Start(srvCtx, fmt.Sprintf("%s:%d", cmdConfig.RPC.Host, cmdConfig.RPC.Port)); err1 != nil {
			once.Do(srvCancel)
		}
	})

	var err2 error
	wg.Go(func() {
		logg.Info("Начало работы сервиса мониторинга")
		err2 = metricsCollector.Run(srvCtx)
		logg.Info("Завершение работы сервиса мониторинга")
	})

	wg.Go(func() {
		time.Sleep(time.Duration(serverTimeout+2) * time.Second)
		pid, _, _ := syscall.Syscall(syscall.SYS_GETPID, 0, 0, 0)
		process, _ := os.FindProcess(int(pid))
		process.Signal(syscall.SIGHUP)
	})

	time.Sleep(2 * time.Second)

	cltCtx, CltCancel := context.WithTimeout(context.Background(), time.Duration(clientTimeout)*time.Second)
	defer CltCancel()

	addr := fmt.Sprintf("%s:%d", cmdConfig.RPC.Host, cmdConfig.RPC.Port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoErrorf(t, err, "Не удалось подключиться к %s: %v", addr, err)

	defer conn.Close()

	client := grpcapi.NewStatsServiceClient(conn)

	req := &grpcapi.StatsRequest{
		Interval:        interval,
		AveragingPeriod: averagingPeriod,
		StatTypes:       statTypes,
	}

	stream, err := client.GetStats(cltCtx, req)
	require.NoErrorf(t, err, "Не удалось получить статистику: %v", err)

	logg.Info("Послан запрос статистики")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cltCtx.Done():
			log.Printf("Прервана работа клиента")
			goto END
		case <-ticker.C:

			resp, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					logg.Info("Сервер завершил поток")
					goto END
				}
				st, ok := status.FromError(err)
				require.Equalf(t, true, ok, "Получена не-gRPC ошибка: %v", err)

				if st.Code() != code {
					require.NoErrorf(t, err, "Ошибка сервера: code=%s, message=%q",
						st.Code(),
						st.Message())
				}
				goto ERR_END
			} else {
				logg.Info("Получена статистика: %+v\n", resp)
			}
		}
	}
END:

ERR_END:
	wg.Wait()

	if err1 != nil {
		require.NoErrorf(t, err1, "Не удалось запустить RPC сервер: %v", err1.Error())
	}
	if err2 != nil {
		require.NoErrorf(t, err2, "Ошибка в работе сервиса мониторинга: %v", err2.Error())
	}
}

//nolint:funlen
func TestIntegration(t *testing.T) {
	cmdConfig := config.DaemonConfig{}
	cmdConfig.RPC.Host = "localhost"
	cmdConfig.RPC.Port = 55555
	cmdConfig.Stats.AveragingPeriodLimit = 100
	cmdConfig.Stats.LoadAverage = true
	cmdConfig.Stats.CPUStats = true
	cmdConfig.Stats.DisksLoad = true
	cmdConfig.Stats.DisksStats = true
	cmdConfig.Logger.Level = interfaces.INFO

	t.Run("TC1 Одновременное завершение клиента и сервера", func(t *testing.T) {
		testCase(t, cmdConfig, 10, 10,
			[]grpcapi.StatType{
				grpcapi.StatType_LOAD_AVERAGE,
				grpcapi.StatType_CPU_STATS,
				grpcapi.StatType_DISKS_LOAD,
				grpcapi.StatType_DISKS_STATS,
			},
			1,
			5,
		)
	})

	cmdConfig.RPC.Port = 55556
	t.Run("TC2 Сервер завершается после обработки запроса клиента", func(t *testing.T) {
		testCase(t, cmdConfig, 7, 5,
			[]grpcapi.StatType{
				grpcapi.StatType_CPU_STATS,
				grpcapi.StatType_DISKS_STATS,
			},
			1,
			3,
		)
	})

	cmdConfig.RPC.Port = 55557

	t.Run("TC3 Одновременное завершение клиента и сервера (сервер завершается по сигналу)", func(t *testing.T) {
		testCase2(t, cmdConfig, 10, 10,
			[]grpcapi.StatType{
				grpcapi.StatType_LOAD_AVERAGE,
				grpcapi.StatType_CPU_STATS,
				grpcapi.StatType_DISKS_LOAD,
				grpcapi.StatType_DISKS_STATS,
			},
			1,
			5,
		)
	})

	cmdConfig.RPC.Port = 55558

	t.Run("TC4 Сервер завершается после обработки запроса клиента (сервер завершается по сигналу)", func(t *testing.T) {
		testCase2(t, cmdConfig, 7, 5,
			[]grpcapi.StatType{
				grpcapi.StatType_LOAD_AVERAGE,
				grpcapi.StatType_DISKS_LOAD,
			},
			1,
			3,
		)
	})

	cmdConfig.RPC.Port = 55559
	cmdConfig.Stats.LoadAverage = true
	cmdConfig.Stats.CPUStats = true
	cmdConfig.Stats.DisksLoad = true
	cmdConfig.Stats.DisksStats = false

	t.Run("TC5 Ошибка в запросе от клиента к серверу: запрос не собираемой статистики", func(t *testing.T) {
		testCase3(t, cmdConfig, 7, 5,
			[]grpcapi.StatType{
				grpcapi.StatType_LOAD_AVERAGE,
				grpcapi.StatType_CPU_STATS,
				grpcapi.StatType_DISKS_LOAD,
				grpcapi.StatType_DISKS_STATS,
			},
			1,
			3,
			codes.InvalidArgument,
		)
	})

	cmdConfig.RPC.Port = 55560
	cmdConfig.Stats.LoadAverage = true
	cmdConfig.Stats.CPUStats = true
	cmdConfig.Stats.DisksLoad = true
	cmdConfig.Stats.DisksStats = false
	cmdConfig.Stats.NetworkTopTalkers = true
	cmdConfig.Stats.NetworkConnStats = false

	t.Run("TC6 Ошибка в запросе от клиента к серверу: запрос не реализованной статистики", func(t *testing.T) {
		testCase3(t, cmdConfig, 7, 5,
			[]grpcapi.StatType{
				grpcapi.StatType_LOAD_AVERAGE,
				grpcapi.StatType_CPU_STATS,
				grpcapi.StatType_DISKS_LOAD,
				grpcapi.StatType_NETWORK_TOP_TALKERS,
			},
			1,
			3,
			codes.InvalidArgument,
		)
	})

	cmdConfig.RPC.Port = 55561
	cmdConfig.Stats.LoadAverage = true
	cmdConfig.Stats.CPUStats = true
	cmdConfig.Stats.DisksLoad = true
	cmdConfig.Stats.DisksStats = true
	cmdConfig.Stats.NetworkTopTalkers = false
	cmdConfig.Stats.NetworkConnStats = false

	t.Run("TC7 Ошибка в запросе от клиента к серверу: неверный параметр averagingPeriod", func(t *testing.T) {
		testCase3(t, cmdConfig, 7, 5,
			[]grpcapi.StatType{
				grpcapi.StatType_LOAD_AVERAGE,
				grpcapi.StatType_CPU_STATS,
				grpcapi.StatType_DISKS_LOAD,
				grpcapi.StatType_DISKS_STATS,
			},
			1,
			101,
			codes.InvalidArgument,
		)
	})

	cmdConfig.RPC.Port = 55562

	t.Run("TC8 Ошибка в запросе от клиента к серверу: неверный параметр averagingPeriod", func(t *testing.T) {
		testCase3(t, cmdConfig, 7, 5,
			[]grpcapi.StatType{
				grpcapi.StatType_LOAD_AVERAGE,
				grpcapi.StatType_CPU_STATS,
				grpcapi.StatType_DISKS_LOAD,
				grpcapi.StatType_DISKS_STATS,
			},
			1,
			0,
			codes.InvalidArgument,
		)
	})

	cmdConfig.RPC.Port = 55563

	t.Run("TC9 Ошибка в запросе от клиента к серверу: неверный параметр interval", func(t *testing.T) {
		testCase3(t, cmdConfig, 7, 5,
			[]grpcapi.StatType{
				grpcapi.StatType_LOAD_AVERAGE,
				grpcapi.StatType_CPU_STATS,
				grpcapi.StatType_DISKS_LOAD,
				grpcapi.StatType_DISKS_STATS,
			},
			0,
			5,
			codes.InvalidArgument,
		)
	})

	cmdConfig.RPC.Port = 55564
	cmdConfig.Stats.LoadAverage = true
	cmdConfig.Stats.CPUStats = true
	cmdConfig.Stats.DisksLoad = true

	t.Run("TC10 сервер завершается раньше клиента", func(t *testing.T) {
		testCase3(t, cmdConfig, 4, 7,
			[]grpcapi.StatType{
				grpcapi.StatType_LOAD_AVERAGE,
				grpcapi.StatType_CPU_STATS,
				grpcapi.StatType_DISKS_LOAD,
			},
			1,
			5,
			codes.Aborted,
		)
	})
}
