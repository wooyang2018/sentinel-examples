package main

import (
	"context"
	"testing"

	pb "github.com/go-kratos/examples/helloworld/helloworld"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestMain1(t *testing.T) {
	// new etcd client
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"127.0.0.1:2379"},
	})
	if err != nil {
		panic(err)
	}
	// new discovery with etcd client
	dis := etcd.New(client)

	endpoint := "discovery:///helloworld"
	// 由于 gRPC 框架的限制，只能使用全局 balancer name 的方式来注入 selector
	selector.SetGlobalSelector(wrr.NewBuilder())

	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(endpoint),
		grpc.WithDiscovery(dis),
		// 通过 grpc.WithFilter 注入路由 Filter
		grpc.WithNodeFilter(filter("test")),
		grpc.WithMiddleware(myMiddleware),
	)
	if err != nil {
		panic(err)
	}

	client2 := pb.NewGreeterClient(conn)
	// 创建一个HelloRequest
	req := &pb.HelloRequest{Name: "World"}

	// 调用服务端的SayHello方法
	reply, err := client2.SayHello(context.Background(), req)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Infof("Greeting: %s", reply.GetMessage())
}
