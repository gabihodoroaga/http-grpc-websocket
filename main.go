package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

	"github.com/gabihodoroaga/http-grpc-websocket/config"
	appgrpc "github.com/gabihodoroaga/http-grpc-websocket/grpc"
	pb "github.com/gabihodoroaga/http-grpc-websocket/grpc/proto"
	apphttp "github.com/gabihodoroaga/http-grpc-websocket/http"
)

func main() {
	logger, err := initLogger()
	if err != nil {
		fmt.Printf("error initializing logger %v", err)
		os.Exit(1)
	}

	logger.Info("main: start")

	config.LoadConfig()

	r := gin.Default()
	apphttp.Setup(r)

	addr := resolveAddress()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Sugar().Errorf("error listen tcp listen on  %s, %v", addr, err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(appgrpc.ServerOptions()...)
	pb.RegisterEchoServer(grpcServer, appgrpc.NewServer())

	mixedHandler := newHTTPAndGRPCMux(r, grpcServer)
	http2Server := &http2.Server{}
	http1Server := &http.Server{Handler: h2c.NewHandler(mixedHandler, http2Server)}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	go func(l net.Listener) {
		logger.Sugar().Infof("httpServer: listen on address %s", addr)
		if err := http1Server.Serve(l); err != nil && err != http.ErrServerClosed {
			logger.Sugar().Errorf("http1Server exit with error %v", err)
			os.Exit(1)
		}
	}(listener)

	<-ctx.Done()
	stop()

	fmt.Println("httpServer: shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := http1Server.Shutdown(ctx); err != nil {
		logger.Sugar().Errorf("forced to shutdown, error %v", err)
		os.Exit(1)
	}
	fmt.Println("httpServer: server exiting...")
}

func resolveAddress() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	return ":8080"
}

func newHTTPAndGRPCMux(httpHand http.Handler, grpcHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("content-type"), "application/grpc") {
			grpcHandler.ServeHTTP(w, r)
			return
		}
		httpHand.ServeHTTP(w, r)
	})
}

func initLogger() (*zap.Logger, error) {
	loggerConfig := zap.NewProductionConfig()

	loggerConfig.Level.SetLevel(resolveLogLevel())
	loggerConfig.EncoderConfig.LevelKey = "severity"
	loggerConfig.EncoderConfig.MessageKey = "message"
	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, err
	}
	zap.ReplaceGlobals(logger)
	return logger, nil
}

func resolveLogLevel() zapcore.Level {
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		var l zapcore.Level
		if err := l.UnmarshalText([]byte(level)); err != nil {
			return zap.InfoLevel
		}
		return l
	}
	return zap.InfoLevel
}
