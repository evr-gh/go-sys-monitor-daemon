package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/interfaces"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/logger"
	grpcapi "github.com/evr-gh/go-sys-monitor-deamon/internal/server/rpc/grpcapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var (
	host              = flag.String("host", "localhost", "Хост серверв")
	port              = flag.String("port", "8001", "Порт сервера")
	logLevelStr       = flag.String("log_level", "INFO", "Уровень логирования: DEBUG, INFO, WARNING или ERROR")
	interval          = flag.Int64("interval", 1, "Интервал между получением статистики")
	averagingPeriod   = flag.Int64("averaging-period", 2, "Период усреднения в секундах")
	loadAverage       = flag.Bool("load_average", false, "Средняя загрузка системы")
	cpuStats          = flag.Bool("cpu_stats", false, "Статистика по ЦПУ")
	disksLoad         = flag.Bool("disks_load", false, "Средняя загрузка дисков")
	disksStats        = flag.Bool("disks_starts", false, "Статистика по дискам")
	networkTopTalkers = flag.Bool("network_top_talkers", false, "Top talkers по сети")
	networkConnStats  = flag.Bool("network_conn_stats", false, "Статистика по сетевым соединениям")
)

// ./client -load-avg=false -disk-usage=false

func main() {
	flag.Parse()

	var logLevel interfaces.LogLevel
	if logLevelStr == nil {
		fmt.Fprintf(os.Stderr, "Не задано значение параметра log_level: %s", *logLevelStr)
	}
	switch *logLevelStr {
	case "DEBUG":
		logLevel = interfaces.DEBUG
	case "INFO":
		logLevel = interfaces.INFO
	case "WARNING":
		logLevel = interfaces.WARNING
	case "ERROR":
		logLevel = interfaces.ERROR
	default:
		fmt.Fprintf(os.Stderr, "Неверное значение параметра log_level: %s", *logLevelStr)
		logLevel = interfaces.INFO
	}
	logg := logger.New(logLevel, os.Stdout)

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer stop()

	addr := fmt.Sprintf("%s:%s", *host, *port)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logg.Error("Не удалось подключиться к %s: %v", addr, err)
		return
	}
	defer conn.Close()

	client := grpcapi.NewStatsServiceClient(conn)

	var statTypes []grpcapi.StatType
	if *loadAverage {
		statTypes = append(statTypes, grpcapi.StatType_LOAD_AVERAGE)
	}
	if *cpuStats {
		statTypes = append(statTypes, grpcapi.StatType_CPU_STATS)
	}
	if *disksLoad {
		statTypes = append(statTypes, grpcapi.StatType_DISKS_LOAD)
	}
	if *disksStats {
		statTypes = append(statTypes, grpcapi.StatType_DISKS_STATS)
	}
	if *networkTopTalkers {
		statTypes = append(statTypes, grpcapi.StatType_NETWORK_TOP_TALKERS)
	}
	if *networkConnStats {
		statTypes = append(statTypes, grpcapi.StatType_NETWORK_CONN_STATS)
	}

	if len(statTypes) == 0 {
		logg.Error("Не выбраны катагории для получения статистики")
		return
	}

	req := &grpcapi.StatsRequest{
		Interval:        *interval,
		AveragingPeriod: *averagingPeriod,
		StatTypes:       statTypes,
	}

	stream, err := client.GetStats(ctx, req)
	if err != nil {
		logg.Error("Не удалось получить статистику: %v", err)
		return
	}
	logg.Info("Послан запрос статистики")

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Прервана работа клиента")
			return
		case <-ticker.C:
			resp, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					logg.Info("Сервер завершил поток")
					return
				}

				st, ok := status.FromError(err)
				if !ok {
					logg.Error("Получена не-gRPC ошибка: %v", err)
					return
				}

				logg.Error(
					"Ошибка сервера: code=%s, message=%q",
					st.Code(),
					st.Message(),
				)

				return
			}

			logg.Info("Получена статистика: %+v\n", resp)
		}
	}
}
