package handler

import (
	"context"
	"errors"
	"io"
	"time"

	"go-micro.dev/v4/logger"

	pb "sentinel-examples/helloworld/proto"
)

type Helloworld struct {
	Name string
}

func (e *Helloworld) Call(ctx context.Context, req *pb.CallRequest, rsp *pb.CallResponse) error {
	rsp.Msg = "Hello " + req.Name + ", I am " + e.Name
	if e.Name[0]%2 == 0 {
		return errors.New("even name is not allowed")
	}
	return nil
}

func (e *Helloworld) ClientStream(ctx context.Context, stream pb.Helloworld_ClientStreamStream) error {
	var count int64
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			logger.Infof("Got %v pings total", count)
			return stream.SendMsg(&pb.ClientStreamResponse{Count: count})
		}
		if err != nil {
			return err
		}
		logger.Infof("Got ping %v", req.Stroke)
		count++
	}
}

func (e *Helloworld) ServerStream(ctx context.Context, req *pb.ServerStreamRequest, stream pb.Helloworld_ServerStreamStream) error {
	logger.Infof("Received Helloworld.ServerStream request: %v", req)
	for i := 0; i < int(req.Count); i++ {
		logger.Infof("Sending %d", i)
		if err := stream.Send(&pb.ServerStreamResponse{
			Count: int64(i),
		}); err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 250)
	}
	return nil
}

func (e *Helloworld) BidiStream(ctx context.Context, stream pb.Helloworld_BidiStreamStream) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		logger.Infof("Got ping %v", req.Stroke)
		if err := stream.Send(&pb.BidiStreamResponse{Stroke: req.Stroke}); err != nil {
			return err
		}
	}
}
