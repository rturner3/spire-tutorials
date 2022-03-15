package proxy

import (
	"fmt"
	"net"
	"os"
	logcommon "delegatedidentity/internal/common/logger"
	"delegatedidentity/proto/hello"

	"github.com/spiffe/spire/pkg/common/peertracker"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewProxy(helloServer hello.HelloServer) *grpc.Server {
	server := grpc.NewServer(
		grpc.Creds(peertracker.NewCredentials()),
	)

	hello.RegisterHelloServer(server, helloServer)
	return server
}

func RunProxy(config Config, server *grpc.Server, logger *zap.Logger) error {
	lis, err := newListener(config.ListenSocketPath, logger)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	defer lis.Close()

	return server.Serve(lis)
}

func newListener(socketPath string, logger *zap.Logger) (net.Listener, error) {
	logrusLogger := logcommon.NewLogrusBridgeLogger(logger)
	os.Remove(socketPath)
	lf := peertracker.ListenerFactory{
		Log: logrusLogger,
	}

	unixAddr, err := net.ResolveUnixAddr("unix", socketPath)
	if err != nil {
		return nil, err
	}

	l, err := lf.ListenUnix(unixAddr.Net, unixAddr)
	if err != nil {
		return nil, fmt.Errorf("create UDS listener: %w", err)
	}

	if err := os.Chmod(unixAddr.String(), os.ModePerm); err != nil {
		return nil, fmt.Errorf("unable to change UDS permissions: %w", err)
	}

	return l, nil
}
