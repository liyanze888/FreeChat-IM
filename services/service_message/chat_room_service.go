package service_message

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	grpc_common "freechat/im/grpc-common"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
	"github.com/liyanze888/funny-core/fn_utils"
	"time"
)

/**
 * 会话处理
 */
func init() {
	fn_factory.BeanFactory.RegisterBean(NewChatRoomControl())
}

type ChatRoomService interface {
	SendMsg(msg *gatewaypb.MessageWrapper, userInfo *grpc_common.GrpcContextInfo)
}

type chatRoomService struct {
	MessageWorker MessageDispatcher  `autowire:""`
	IdUtils       fn_utils.IdUtils   `autowire:""`
	ActionMessage AsyncActionMessage `autowire:""`
}

func (c *chatRoomService) SendMsg(msg *gatewaypb.MessageWrapper, userInfo *grpc_common.GrpcContextInfo) {
	payloads := msg.GetPayloads()
	// 是否combine
	for _, payload := range payloads {
		id, err := c.IdUtils.NextID()
		payload.Sender = userInfo.UserId
		payload.CreateTime = time.Now().UnixMilli()
		if err != nil {
			fn_log.Printf("%v", err)
			continue
		}

		//全局ID
		payload.ServerId = int64(id)
		if msgHolder, err := c.MessageWorker.Dispatch(payload, userInfo); err != nil {
			fn_log.Printf("Dispatch = %v", err)
		} else {
			fn_log.Printf("SendMsg = %v", err)

			cells := make([]*gatewaypb.MessageCell, 0, len(msgHolder.Payloads))
			for _, holderPaylod := range msgHolder.Payloads {
				cells = append(cells, holderPaylod.Message)
			}
			payload.Items = cells
			action := ActionMessage{
				userInfo: userInfo,
				sendTo:   msgHolder.SendTo,
				Save:     msgHolder.Save,
				Push:     msgHolder.Push,
				payload:  payload,
			}
			c.ActionMessage.async(action)
		}
	}
}

func NewChatRoomControl() ChatRoomService {
	return &chatRoomService{}
}
