package grpc

import (
	"context"
	"net"

	"gitlab.com/marsskom/burro/internal/broker"
	"gitlab.com/marsskom/burro/internal/config"
	"gitlab.com/marsskom/burro/internal/logger"
	pt "gitlab.com/marsskom/burro/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type burroServer struct {
	hub *broker.Hub

	pt.UnimplementedBurroServer
}

func (bs *burroServer) Ping(ctx context.Context, req *pt.PingRequest) (*pt.PingResponse, error) {
	return &pt.PingResponse{
		Message: "pong",
	}, nil
}

func (bs *burroServer) Subscribe(req *pt.SubscribeRequest, stream grpc.ServerStreamingServer[pt.Event]) error {
	ch := bs.hub.Subscribe()
	defer bs.hub.Unsubscribe(ch)

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()

		case e, ok := <-ch:
			if !ok {
				return nil
			}

			if err := stream.Send(brokerEventToProtoEvent(e)); err != nil {
				return err
			}
		}
	}
}

type ServerWrapper struct {
	enabled       bool
	listen        string
	debug         bool
	silentFailure bool

	bs *burroServer

	Server *grpc.Server
}

func NewServerWrapper(cfg *config.Config, hub *broker.Hub) *ServerWrapper {
	return &ServerWrapper{
		enabled:       cfg.GRPC.Enabled,
		listen:        cfg.GRPC.Listen,
		debug:         cfg.GRPC.Debug,
		silentFailure: cfg.Proxy.ZeroConfigurationMode,

		bs: &burroServer{
			hub: hub,
		},
	}
}

func (s *ServerWrapper) Start(errCh chan<- error) {
	if !s.enabled {
		logger.Info("gRPC server is disabled")
		return
	}

	lis, err := net.Listen("tcp", s.listen)
	if err != nil {
		s.handleError(err, errCh)
		return
	}

	s.Server = grpc.NewServer()

	pt.RegisterBurroServer(s.Server, s.bs)

	if s.debug {
		reflection.Register(s.Server)
	}

	go func() {
		logger.Info("gRPC server is running", "address", s.listen)

		err := s.Server.Serve(lis)
		if err != nil {
			s.handleError(err, errCh)
		}
	}()
}

func (s *ServerWrapper) handleError(err error, errCh chan<- error) {
	if s.silentFailure {
		logger.Warn("gRPC silently failed", "err", err)
		return
	}

	errCh <- err
}

func (s *ServerWrapper) Stop(ctx context.Context) {
	if s.Server == nil {
		return
	}

	done := make(chan struct{})

	go func() {
		s.Server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		s.Server.Stop()
	}

	logger.Info("gRPC server stopped")
}
