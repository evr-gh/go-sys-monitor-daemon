package rpcserver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	interfaces "github.com/evr-gh/go-sys-monitor-deamon/internal/interfaces"
	"github.com/evr-gh/go-sys-monitor-deamon/internal/server/rpc/grpcapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type RPCServer struct {
	grpcapi.UnimplementedStatsServiceServer
	logger interfaces.Logger
	cltr   interfaces.Collector
	ctx    context.Context
	mu     sync.Mutex
	server *grpc.Server
}

func NewRPCServer(logger interfaces.Logger, cltr interfaces.Collector) *RPCServer {
	return &RPCServer{
		logger: logger,
		cltr:   cltr,
		ctx:    context.Background(),
	}
}

func LoggingdUnaryInterceptor(
	logger interfaces.Logger,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		logger.Debug(
			"Получено gRPC сообщение: method=%q request={%v}",
			info.FullMethod,
			req,
		)

		start := time.Now()
		res, err := handler(ctx, req)
		duration := time.Since(start)
		code := status.Code(err)

		if err != nil {
			logger.Info("Выполнение метода: method=%q out={%v} code=%s duration=%s error=%v",
				info.FullMethod, res, code, duration, err)

			return res, err
		}

		logger.Info("Выполнение метода: method=%q out={%v} code=%s duration=%s",
			info.FullMethod, res, code, duration)
		return res, nil
	}
}

type loggingServerStream struct {
	grpc.ServerStream

	logger interfaces.Logger
	method string

	received atomic.Uint64
	sent     atomic.Uint64
}

func (s *loggingServerStream) RecvMsg(message any) error {
	err := s.ServerStream.RecvMsg(message)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			s.logger.Error(
				"Ошибка чтения gRPC сообщения: method=%q error=%v",
				s.method,
				err,
			)
		}

		return err
	}

	s.received.Add(1)

	s.logger.Debug(
		"Получено gRPC сообщение: method=%q request={%v}",
		s.method,
		message,
	)

	return nil
}

func (s *loggingServerStream) SendMsg(message any) error {
	if err := s.ServerStream.SendMsg(message); err != nil {
		s.logger.Error(
			"Ошибка отправки gRPC сообщения: method=%q error=%v",
			s.method,
			err,
		)

		return err
	}

	s.sent.Add(1)

	s.logger.Debug(
		"Отправлено gRPC сообщение: method=%q response={%v}",
		s.method,
		message,
	)

	return nil
}

func LoggingStreamInterceptor(
	logger interfaces.Logger,
) grpc.StreamServerInterceptor {
	return func(
		srv any,
		stream grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		logger.Info(
			"Начало gRPC потока: method=%q client_stream=%t server_stream=%t",
			info.FullMethod,
			info.IsClientStream,
			info.IsServerStream,
		)

		wrappedStream := &loggingServerStream{
			ServerStream: stream,
			logger:       logger,
			method:       info.FullMethod,
		}

		start := time.Now()

		err := handler(srv, wrappedStream)

		duration := time.Since(start)
		code := status.Code(err)

		if err != nil {
			logger.Info(
				"Конец gRPC потока: method=%q code=%s duration=%s error=%v",
				info.FullMethod,
				code,
				duration,
				err,
			)

			return err
		}

		logger.Info(
			"Конец gRPC потока: method=%q code=%s duration=%s",
			info.FullMethod,
			code,
			duration,
		)

		return nil
	}
}

func (s *RPCServer) getServer() *grpc.Server {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.server
}

func (s *RPCServer) Start(ctx context.Context, address string) error {
	listenConfig := net.ListenConfig{}
	listener, err := listenConfig.Listen(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf(
			"не удалось открыть gRPC listener %q: %w",
			address,
			err,
		)
	}
	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(LoggingdUnaryInterceptor(s.logger)),
		grpc.StreamInterceptor(LoggingStreamInterceptor(s.logger)),
	)
	grpcapi.RegisterStatsServiceServer(gRPCServer, s)
	s.mu.Lock()

	if s.server != nil {
		s.mu.Unlock()
		_ = listener.Close()

		return errors.New("gRPC сервер уже запущен")
	}

	s.server = gRPCServer
	s.ctx = ctx
	s.mu.Unlock()

	// Очищаем поле, когда Serve завершится.
	defer func() {
		s.mu.Lock()

		// Проверка нужна, чтобы случайно не очистить ссылку
		// на другой экземпляр сервера.
		if s.server == gRPCServer {
			s.server = nil
		}

		s.mu.Unlock()
	}()

	s.logger.Info("Запуск gRPC сервера: address=%q", address)

	err = gRPCServer.Serve(listener)
	if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("ошибка работы gRPC сервера: %w", err)
	}

	return nil
}

func (s *RPCServer) Stop() {
	server := s.getServer()

	if server == nil {
		s.logger.Info("gRPC сервер не запущен")
		return
	}

	s.logger.Info("Принудительный останов gRPC сервера")
	server.Stop()
}

func (s *RPCServer) GracefulStop() {
	server := s.getServer()

	if server == nil {
		s.logger.Info("gRPC сервер не запущен")
		return
	}

	s.logger.Info("Штатный останов gRPC сервера")
	server.GracefulStop()
}

func (s *RPCServer) GetStats(req *grpcapi.StatsRequest, stream grpcapi.StatsService_GetStatsServer) error {
	peer, ok := peer.FromContext(stream.Context())
	clientAddr := "unknown"
	if ok {
		clientAddr = peer.Addr.String()
	}
	defer func() {
		if r := recover(); r != nil {
			s.logger.Error(fmt.Sprintf("Критическая ошибка в GetStats (%s): %v", clientAddr, r))
		}
		s.logger.Info(fmt.Sprintf("Клиент %s отсоединился", clientAddr))
	}()

	s.logger.Info(fmt.Sprintf("Новый запрос от %s: interval=%d, averaging_period=%d, stat_types=%v",
		clientAddr, req.Interval, req.AveragingPeriod, req.StatTypes))

	if len(req.StatTypes) == 0 {
		s.logger.Error("Не заданы категории статистики для получения")
		return status.Errorf(codes.InvalidArgument, "не заданы категории статистики для получения")
	}

	if req.Interval < 1 {
		s.logger.Error(fmt.Sprintf("Интервал между получением статистики (%d) меньше 1", req.Interval))
		return status.Errorf(codes.InvalidArgument, "Интервал между получением статистики должен быть >=1")
	}

	if req.AveragingPeriod < 1 {
		s.logger.Error(fmt.Sprintf("Период усреднения (%d) меньше 1", req.AveragingPeriod))
		return status.Errorf(codes.InvalidArgument, "период усреднения должен быть >=1")
	}

	err := s.cltr.RegisterRequest(req.AveragingPeriod, req.StatTypes)
	if err != nil {
		s.logger.Error("Ошибка в запросе: %v", err)
		return status.Errorf(codes.InvalidArgument, "Ошибка в запросе: %v", err.Error())
	}
	defer s.cltr.OnRequestCancel(req.AveragingPeriod, req.StatTypes)

	if req.AveragingPeriod > req.Interval {
		for {
			select {
			case <-s.ctx.Done():
				s.logger.Info(fmt.Sprintf("Запрос от %s отменен по завершению работы сервера", clientAddr))
				return status.Errorf(codes.Aborted, "завершение работы сервера")
			case <-stream.Context().Done():
				s.logger.Info(fmt.Sprintf("Запрос отменен клиентом %s", clientAddr))
				return status.Errorf(codes.Aborted, "Запрос отменен по запросу клиента")
			case <-time.After(time.Duration(req.AveragingPeriod-req.Interval) * time.Second):
				goto END
			}
		}
	END:
	}

	sendTicker := time.NewTicker(time.Duration(req.Interval) * time.Second)
	defer sendTicker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.logger.Info(fmt.Sprintf("Запрос от %s отменен по завершению работы сервера", clientAddr))
			return status.Errorf(codes.Aborted, "завершение работы сервера")
		case <-stream.Context().Done():
			s.logger.Info(fmt.Sprintf("Запрос отменен клиентом %s", clientAddr))
			return status.Errorf(codes.Aborted, "Запрос отменен по запросу клиента")
		case <-sendTicker.C:
			response := s.cltr.PrepareResponse(req.AveragingPeriod, req.StatTypes)
			if err := stream.Send(response); err != nil {
				s.logger.Error(fmt.Sprintf("Не удалось передать статистику клтиенту (%s): %v", clientAddr, err))
				return fmt.Errorf("не удалось передать статистику клтиенту: %w", err)
			}
		}
	}
}
