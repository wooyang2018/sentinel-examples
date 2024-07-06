package client

import (
	"context"
	"fmt"

	sentinelApi "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/selector"
)

type clientWrapper struct {
	client.Client
	Opts []Option
}

func (c *clientWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	resourceName := req.Method()
	options := evaluateOptions(c.Opts)

	if options.clientResourceExtract != nil {
		resourceName = options.clientResourceExtract(ctx, req)
	}

	entry, blockErr := sentinelApi.Entry(
		resourceName,
		sentinelApi.WithResourceType(base.ResTypeRPC),
		sentinelApi.WithTrafficType(base.Outbound),
	)

	if blockErr != nil {
		if options.clientBlockFallback != nil {
			return options.clientBlockFallback(ctx, req, blockErr)
		}
		return blockErr
	}
	defer entry.Exit()

	opt1 := client.WithSelectOption(selector.WithFilter(
		func(old []*registry.Service) []*registry.Service {
			for _, service := range old {
				for _, ep := range service.Nodes {
					fmt.Println("opt1", ep.Id)
				}
			}
			return old
		},
	))
	opt2 := client.WithCallWrapper(func(f1 client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
			fmt.Println("opt2", node.Id)
			return f1(ctx, node, req, rsp, opts)
		}
	})
	opts = append(opts, opt1, opt2)
	err := c.Client.Call(ctx, req, rsp, opts...)
	if err != nil {
		sentinelApi.TraceError(entry, err)
	}

	return err
}

func (c *clientWrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	options := evaluateOptions(c.Opts)
	resourceName := req.Method()

	if options.streamClientResourceExtract != nil {
		resourceName = options.streamClientResourceExtract(ctx, req)
	}

	entry, blockErr := sentinelApi.Entry(
		resourceName,
		sentinelApi.WithResourceType(base.ResTypeRPC),
		sentinelApi.WithTrafficType(base.Outbound),
	)

	if blockErr != nil {
		if options.streamClientBlockFallback != nil {
			return options.streamClientBlockFallback(ctx, req, blockErr)
		}
		return nil, blockErr
	}
	defer entry.Exit()

	stream, err := c.Client.Stream(ctx, req, opts...)
	if err != nil {
		sentinelApi.TraceError(entry, err)
	}

	return stream, err
}

// NewClientWrapper returns a sentinel client Wrapper.
func NewClientWrapper(opts ...Option) client.Wrapper {
	return func(c client.Client) client.Client {
		return &clientWrapper{c, opts}
	}
}
