package fc_im_grpc_server

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_grpc"
	"sync"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewFcImGrpcServer())
	fn_grpc.GrpcBeanFactory.Register(gatewaypb.RegisterImServiceServer)
}

type fcImGrpcServer struct {
	gatewaypb.UnsafeImServiceServer
	holders map[int]map[int64]ConnectHolder
	Worker  MessageWorker `autowire:""`
}

// Connect 连接
func (im *fcImGrpcServer) Connect(server gatewaypb.ImService_ConnectServer) error {
	wg := sync.WaitGroup{}
	holder := NewConnectHolder(&wg, server, im.Worker)
	holder.Start()
	wg.Wait()
	return nil
}

func NewFcImGrpcServer() gatewaypb.ImServiceServer {
	return &fcImGrpcServer{}
}
