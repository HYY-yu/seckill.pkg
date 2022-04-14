package metrics

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

var (
	// GRPCMetrics Create some standard server metrics.
	GRPCMetrics *grpc_prometheus.ServerMetrics
)

func InitGrpcMetrics(grpcServer *grpc.Server) {
	// Register standard server metrics and customized metrics to registry.
	GRPCMetrics = grpc_prometheus.NewServerMetrics()
	prometheus.MustRegister(GRPCMetrics)

}

// SetToGrpcServer 注意，设置前，需要先注册拦截器
//	grpcServer := grpc.NewServer(
//		grpc.StreamInterceptor(GRPCMetrics.StreamServerInterceptor()),
//		grpc.UnaryInterceptor(GRPCMetrics.UnaryServerInterceptor()),
//	)
func SetToGrpcServer(grpcServer *grpc.Server) {
	GRPCMetrics.InitializeMetrics(grpcServer)
}
