package service_message

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	messagepb "freechat/im/generated/grpc/im/gateway/message"
	messageType "freechat/im/generated/grpc/im/message"
	"freechat/im/subscribe"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewConnectMessageWorker())
}

type connectMessageWorker struct {
	workerType int
}

func (c *connectMessageWorker) work(msg *gatewaypb.MessageCell, user *subscribe.UserContext) (*MessageHolder, error) {
	fn_log.Printf("%v", msg)

	holder := &MessageHolder{
		Message: &gatewaypb.MessageCell{
			ChatId: msg.GetChatId(),
			MessageComment: &gatewaypb.MessageCell_ConnectMessage{
				ConnectMessage: &messagepb.ConnectMessage{
					ConnectMessageType: messagepb.ConnectMessageType_ConnectMessageTypeBeatHeart,
					ConnectMessage:     &messagepb.ConnectMessage_BeatheartMessage{},
				},
			},
			Show: true,
		},
		Push: false,
		Save: true,
	}
	return holder, nil
}

func NewConnectMessageWorker() MessageTypeWorker {
	return &connectMessageWorker{
		workerType: int(messageType.MessageType_MessageTypeConnect),
	}
}
