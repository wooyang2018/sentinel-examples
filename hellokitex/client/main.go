package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/rpcinfo/remoteinfo"
	etcd "github.com/kitex-contrib/registry-etcd"
)

func main() {
	r, err := etcd.NewEtcdResolver([]string{"127.0.0.1:2379"})
	if err != nil {
		log.Fatal(err)
	}
	client, err := hello.NewClient("example.hello",
		client.WithResolver(r),
		client.WithMiddleware(myMiddleware),
	)
	if err != nil {
		log.Fatal(err)
	}

	req := &api.Request{Message: "Kitex"}
	// callopt.WithHostPort("localhost:8888")
	resp, err := client.Echo(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(resp)
	time.Sleep(time.Second)
}

func myMiddleware(src endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req, resp interface{}) (err error) {
		err = src(ctx, req, resp)
		rpcInfo := rpcinfo.GetRPCInfo(ctx)
		dest := rpcInfo.To()
		if dest == nil {
			return kerrors.ErrNoDestService
		}
		remote := remoteinfo.AsRemoteInfo(dest)
		if remote == nil {
			err := fmt.Errorf("unsupported target EndpointInfo type: %T", dest)
			return kerrors.ErrInternalException.WithCause(err)
		}
		fmt.Printf("%+v\n", remote.Address())
		return err
	}
}
