package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"github.com/stretchr/testify/assert"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/selector"

	proto "helloworld/proto"
)

const FakeErrorMsg = "fake error for testing"

func TestClientLimiter(t *testing.T) {
	//r := consul.NewRegistry(
	//	registry.Addrs("localhost:8500"),
	//)
	r := etcd.NewRegistry(
		registry.Addrs("localhost:2379"),
	)
	s := selector.NewSelector(
		selector.Registry(r),
		selector.SetStrategy(selector.RoundRobin),
	)

	c := client.NewClient(
		// set the selector
		client.Selector(s),
		// add the breaker wrapper
		client.Wrap(NewClientWrapper(
			// add custom fallback function to return a fake error for assertion
			WithClientBlockFallback(
				func(ctx context.Context, request client.Request, blockError *base.BlockError) error {
					return errors.New(FakeErrorMsg)
				}),
		)),
	)

	req := c.NewRequest("helloworld", "Helloworld.Call", &proto.CallRequest{Name: "Bob"}, client.WithContentType("application/json"))

	err := sentinel.InitDefault()
	if err != nil {
		log.Fatal(err)
	}

	rsp := &proto.CallResponse{}
	t.Run("success", func(t *testing.T) {
		var _, err = flow.LoadRules([]*flow.Rule{
			{
				Resource:               req.Method(),
				Threshold:              1.0,
				TokenCalculateStrategy: flow.Direct,
				ControlBehavior:        flow.Reject,
			},
		})
		assert.Nil(t, err)

		err = c.Call(context.TODO(), req, rsp)
		fmt.Println(rsp)
		assert.Nil(t, err)
		assert.True(t, util.Float64Equals(1.0, stat.GetResourceNode(req.Method()).GetQPS(base.MetricEventPass)))

		t.Run("second fail", func(t *testing.T) {
			err := c.Call(context.TODO(), req, rsp)
			assert.EqualError(t, err, FakeErrorMsg)
			assert.True(t, util.Float64Equals(1.0, stat.GetResourceNode(req.Method()).GetQPS(base.MetricEventPass)))
		})
	})
}
