package service_message

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	"freechat/im/subscribe"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
)

/**
 * 会话处理
 */
func init() {
	fn_factory.BeanFactory.RegisterBean(NewChatRoomControl())
}

type ChatRoomService interface {
	SendMsg(msg *gatewaypb.MessageWrapper, userStream *subscribe.UserContext)
}

type chatRoomService struct {
	MessageWorker MessageDispatcher `autowire:""`
}

func (c *chatRoomService) SendMsg(msg *gatewaypb.MessageWrapper, user *subscribe.UserContext) {
	if msg, err := c.MessageWorker.Dispatch(msg, user); err != nil {
		fn_log.Printf("Dispatch = %v", err)
	} else {
		err := user.Server.Send(msg)
		fn_log.Printf("SendMsg = %v", err)
	}
}

func NewChatRoomControl() ChatRoomService {
	return &chatRoomService{}
}
