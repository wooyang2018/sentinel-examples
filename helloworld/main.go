package main

import (
	"github.com/go-micro/plugins/v4/registry/etcd"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/registry"

	"helloworld/handler"
	pb "helloworld/proto"
)

var (
	service = "helloworld"
	version = "latest"
)

func main() {
	//consulReg := consul.NewRegistry(
	//	registry.Addrs("localhost:8500"),
	//)
	etcdReg := etcd.NewRegistry(
		registry.Addrs("localhost:2379"),
	)

	// Create service
	srv := micro.NewService()
	srv.Init(
		micro.Name(service),
		micro.Version(version),
		micro.Registry(etcdReg),
	)

	// Register handler
	if err := pb.RegisterHelloworldHandler(srv.Server(), &handler.Helloworld{Name: srv.Server().Options().Id}); err != nil {
		logger.Fatal(err)
	}
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
