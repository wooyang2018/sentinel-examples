package main

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/selector"
)

func filter(region string) selector.NodeFilter {
	return func(ctx context.Context, nodes []selector.Node) []selector.Node {
		if v, ok := metadata.FromClientContext(ctx); ok {
			region = v.Get("region") // if a region is specified in the request, use specified in the request
		}

		newNodes := make([]selector.Node, 0, len(nodes))
		for _, node := range nodes {
			if node.Metadata()["region"] == region {
				newNodes = append(newNodes, node)
			}
		}

		if len(newNodes) != 0 {
			return newNodes
		}

		return nodes
	}
}

func myMiddleware(src middleware.Handler) middleware.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		res, err := src(ctx, req)
		p, ok := selector.FromPeerContext(ctx)
		if ok {
			log.Infof("XXXXXXXXXXXXX: %+v", p.Node.Address())
		}
		return res, err
	}
}
