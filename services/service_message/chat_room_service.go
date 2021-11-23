package service_message

import (
	"context"
	"fmt"
	"freechat/im/fc_constant"
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	"freechat/im/rabbit"
	"freechat/im/subscribe"
	"github.com/go-redis/redis/v8"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
	"github.com/liyanze888/funny-core/fn_utils"
	"google.golang.org/protobuf/proto"
	"log"
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
	MessageWorker MessageDispatcher     `autowire:""`
	MqClient      rabbit.RabbitmqClient `autowire:""`
	IdUtils       fn_utils.IdUtils      `autowire:""`
	Rdb           redis.UniversalClient `autowire:""`
}

func (c *chatRoomService) SendMsg(msg *gatewaypb.MessageWrapper, user *subscribe.UserContext) {
	payloads := msg.GetPayloads()
	// 是否combine
	for _, payload := range payloads {
		id, err := c.IdUtils.NextID()
		if err != nil {
			log.Printf("%v", err)
			continue
		}
		//全局ID
		payload.ServerId = int64(id)
		if msgHolder, err := c.MessageWorker.Dispatch(payload, user); err != nil {
			fn_log.Printf("Dispatch = %v", err)
		} else {
			fn_log.Printf("SendMsg = %v", err)

			cells := make([]*gatewaypb.MessageCell, 0, len(msgHolder.Payloads))
			for _, holderPaylod := range msgHolder.Payloads {
				cells = append(cells, holderPaylod.Message)
			}
			payload.Items = cells
			//其实就是消息谁可见
			sendTo := msgHolder.SendTo
			if len(msgHolder.SendTo) == 0 {
				// get all
				sendTo = []int64{user.UserInfo.UserId}
			}
			content, err := proto.Marshal(payload)

			if err != nil {
				fn_log.Printf("Publish message -> marshal error  message Content = %v  error = %v ", content, err)
			}

			go func() {
				for _, userId := range sendTo {
					sendNum := c.Publish(content, userId)
					if sendNum == 0 && msgHolder.Push {
						//push
					}
				}
			}()

			go func() {
				if msgHolder.Save {
					if err = c.MqClient.PublishExchange(fc_constant.ImDbMessageSaveExchange, fc_constant.ImDbMessageSaveKey, content); err != nil {
						fn_log.Printf("%v", err)
					} else {
						//retry
					}
				}
			}()
		}
	}
}

func (c *chatRoomService) Publish(content []byte, userId int64) int64 {
	//todo 加一个chain 批次处理消息
	pubChannel := fmt.Sprintf(subscribe.UserSubscibeChannelTemplate, userId)
	publish := c.Rdb.Publish(context.Background(), pubChannel, content)
	if publish.Err() != nil {
		fn_log.Printf("%v", publish.Err())
	} else {
		result, _ := publish.Result()
		return result
	}
	return 0
}

func (c *chatRoomService) PostInitilization() {
	go func() {
		queue, err := c.MqClient.ListenExchangeWithQueue(fc_constant.ImDbMessageSaveExchange, fc_constant.ImDbMessageSaveQueue, fc_constant.ImDbMessageSaveKey)
		if err != nil {
			panic(err)
		}

		for msg := range queue {
			fn_log.Printf("%v", msg)
			err = c.MqClient.Ack(msg.DeliveryTag)
			if err != nil {
				fn_log.Printf("ack error %v", err)
			}
		}
	}()
}

func NewChatRoomControl() ChatRoomService {
	return &chatRoomService{}
}
