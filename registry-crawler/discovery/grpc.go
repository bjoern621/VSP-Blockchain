package discovery

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	appGRPCConnMu   sync.Mutex
	appGRPCConn     *grpc.ClientConn
	appGRPCConnAddr string
)

// DialAppGRPC establishes a gRPC connection to the app service.
// Repeated calls with the same address will reuse the existing cached connection.
func DialAppGRPC(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	_ = ctx

	appGRPCConnMu.Lock()
	defer appGRPCConnMu.Unlock()

	if appGRPCConn != nil {
		if appGRPCConnAddr != addr {
			return nil, fmt.Errorf("app grpc connection already initialized for %q; cannot reuse for %q", appGRPCConnAddr, addr)
		}
		return appGRPCConn, nil
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	appGRPCConn = conn
	appGRPCConnAddr = addr
	return appGRPCConn, nil
}
