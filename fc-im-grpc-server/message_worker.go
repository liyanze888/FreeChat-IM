package fc_im_grpc_server

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	messagepb "freechat/im/generated/grpc/im/gateway/message"
	message2 "freechat/im/generated/grpc/im/message"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewMessageWorker())
}

type MessageWorker interface {
	Work(message *gatewaypb.MessageWrapper, info *ConnectHolder)
}

type messageWorker struct {
}

func (m messageWorker) Work(message *gatewaypb.MessageWrapper, info *ConnectHolder) {
	fn_log.Printf("msg = %v \n info =%v", message, info)
	info.Server.SendMsg(&gatewaypb.MessageWrapper{
		MsgType: message2.MessageType_MessageTypeConnect,
		MessageComment: &gatewaypb.MessageWrapper_ConnectMessage{
			ConnectMessage: &messagepb.ConnectMessage{
				ConnectMessageType: messagepb.ConnectMessageType_ConnectMessageTypeBeatHeart,
				ConnectMessage:     &messagepb.ConnectMessage_BeatheartMessage{},
			},
		},
	})
}

func NewMessageWorker() MessageWorker {
	return &messageWorker{}
}
