package metrics

import (
	"strings"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

var (
	// GRPCMetrics Create some standard server metrics.
	GRPCMetrics *grpc_prometheus.ServerMetrics
)

func InitGrpcMetrics() error {
	// Register standard server metrics and customized metrics to registry.
	GRPCMetrics = grpc_prometheus.NewServerMetrics()
	// 不能使用 MustRegister ,因为库本身注册过了
	err := prometheus.Register(GRPCMetrics)
	if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
		if strings.Contains(err.Error(), " already exists with the same fully-qualified name and const label values") {
			return nil
		}
		return err
	}
	return nil
}

// SetToGrpcServer 注意，设置前，需要先注册拦截器
//	grpcServer := grpc.NewServer(
//		grpc.StreamInterceptor(GRPCMetrics.StreamServerInterceptor()),
//		grpc.UnaryInterceptor(GRPCMetrics.UnaryServerInterceptor()),
//	)
func SetToGrpcServer(grpcServer *grpc.Server) {
	GRPCMetrics.InitializeMetrics(grpcServer)
}
