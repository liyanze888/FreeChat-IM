package fc_im_grpc_server

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	"freechat/im/services/service_message"
	"freechat/im/subscribe"
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
	holders   map[int]map[int64]*service_message.ConnectHolder
	anonymous map[string]map[int64]*service_message.ConnectHolder
	Worker    service_message.MessageDispatcher  `autowire:""`
	MsFactory *subscribe.MessageSubscribeFactory `autowire:""`
}

// Connect 连接
func (im fcImGrpcServer) Connect(server gatewaypb.ImService_ConnectServer) error {
	wg := sync.WaitGroup{}
	holder, err := service_message.NewConnectHolder(&wg, server, im.Worker, im.MsFactory)
	if err != nil {
		return err
	}
	holder.Start()
	wg.Wait()
	return nil
}

func NewFcImGrpcServer() gatewaypb.ImServiceServer {
	return &fcImGrpcServer{
		holders:   make(map[int]map[int64]*service_message.ConnectHolder),
		anonymous: make(map[string]map[int64]*service_message.ConnectHolder),
	}
}
