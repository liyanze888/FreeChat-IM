package fc_im_grpc_server

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	"freechat/im/services/service_message"
	"freechat/im/subscribe"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_grpc"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewFcImGrpcServer())
	fn_grpc.GrpcBeanFactory.Register(gatewaypb.RegisterImServiceServer)
}

type fcImGrpcServer struct {
	gatewaypb.UnsafeImServiceServer
	MsFactory *subscribe.UserContextFactory      `autowire:""`
	UhFactory *service_message.UserStreamFactory `autowire:""`
}

// Connect 连接
func (im fcImGrpcServer) Connect(server gatewaypb.ImService_ConnectServer) error {
	err := im.UhFactory.StartNewUserStream(server)
	if err != nil {
		return err
	}
	return nil
}

func NewFcImGrpcServer() gatewaypb.ImServiceServer {
	return &fcImGrpcServer{}
}
