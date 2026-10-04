package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/evr-gh/go-sys-monitor-deamon/internal/collector"
	logger "github.com/evr-gh/go-sys-monitor-deamon/internal/logger"
	rpcServer "github.com/evr-gh/go-sys-monitor-deamon/internal/server/rpc"
	"github.com/spf13/pflag"
)

var configFile string

func init() {
	pflag.StringVar(&configFile, "config", "/etc/go-sys-monitor-daemon/config.yaml", "Конфигурационный файл")
}

func main() {
	pflag.Parse()

	if pflag.Arg(0) == "version" {
		printVersion()
		return
	}

	cmdConfig, err := readConfig(configFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var logFile *os.File
	var logg *logger.Logger

	if cmdConfig.Logger.File != "" {
		logFile, err = os.Create(cmdConfig.Logger.File)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Не удалось создать файл для сохранения лога: %v", err)
			os.Exit(1)
		}
		defer logFile.Close()
		logg = logger.New(cmdConfig.Logger.Level, logFile)
	} else {
		logg = logger.New(cmdConfig.Logger.Level, os.Stdout)
	}

	metricsCollector := collector.NewMetricsCollector(logg, cmdConfig.Stats)

	rpcServer := rpcServer.NewRPCServer(logg, metricsCollector)

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer stop()

	wg := sync.WaitGroup{}

	var once sync.Once

	wg.Go(func() {
		<-ctx.Done()
		rpcServer.GracefulStop()
	})

	wg.Go(func() {
		if err := rpcServer.Start(ctx, fmt.Sprintf("%s:%d", cmdConfig.RPC.Host, cmdConfig.RPC.Port)); err != nil {
			logg.Error("Не удалось запустить RPC сервер: %v", err.Error())
			once.Do(stop)
		}
	})

	logg.Info("Начало работы сервиса мониторинга")
	err = metricsCollector.Run(ctx)
	if err != nil {
		logg.Info("Ошибка в работе сервиса мониторинга: %v", err.Error())
	}
	logg.Info("Завершение работы сервиса мониторинга")
	wg.Wait()
}
