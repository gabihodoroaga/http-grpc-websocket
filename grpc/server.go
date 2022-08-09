package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/gabihodoroaga/http-grpc-websocket/config"
	pb "github.com/gabihodoroaga/http-grpc-websocket/grpc/proto"
)

var _ pb.EchoServer = (*Server)(nil)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

// ServerOptions returns the grpc interceptors and other server options
func ServerOptions() []grpc.ServerOption {
	opts := []grpc.ServerOption{}
	if len(config.GetConfig().AuthServiceAccounts) > 0 {
		interceptor := newAuthInterceptor(config.GetConfig().AuthServiceAccounts)
		opts = append(opts, grpc.UnaryInterceptor(interceptor.unary()))
	}
	return opts
}

// Ping implements proto.EchoServer
func (*Server) Ping(context.Context, *pb.PingRequest) (*pb.PongResult, error) {
	return &pb.PongResult{Message: "pong"}, nil
}
