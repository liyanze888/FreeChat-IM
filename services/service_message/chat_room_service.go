package service_message

import (
	"fmt"
	"freechat/im/fc_constant"
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	grpc_common "freechat/im/grpc-common"
	"freechat/im/rabbit"
	"freechat/im/repository"
	"freechat/im/subscribe"
	"github.com/go-redis/redis/v8"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
	"github.com/liyanze888/funny-core/fn_utils"
	"google.golang.org/protobuf/proto"
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
	MessageWorker MessageDispatcher            `autowire:""`
	MqClient      rabbit.RabbitmqClient        `autowire:""`
	IdUtils       fn_utils.IdUtils             `autowire:""`
	MessageRepo   repository.MessageRepository `autowire:""`
	Rdb           redis.UniversalClient        `autowire:""`
	Pubsub        *subscribe.PubSubManager     `autowire:""`
	saveChn       chan ActionMessage
	pubChn        chan ActionMessage
}

type ActionMessage struct {
	holder   *WorkMessageHolder
	payload  *gatewaypb.MessagePayload
	userInfo *grpc_common.GrpcContextInfo
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
			//其实就是消息谁可见
			go func() {
				action := ActionMessage{
					userInfo: userInfo,
					holder:   msgHolder,
					payload:  payload,
				}
				c.pubChn <- action
				c.saveChn <- action
			}()
		}
	}
}

func (c *chatRoomService) asyncPubMsg() {
	go func() {
		for holder := range c.pubChn {
			start := time.Now()
			sendTo := holder.holder.SendTo
			if len(holder.holder.SendTo) == 0 {
				// get all by conversation id
				sendTo = []int64{holder.userInfo.UserId}
			}
			content, err := proto.Marshal(holder.payload)
			if err != nil {
				fn_log.Printf("Publish message -> marshal error  message Content = %v  error = %v ", content, err)
				continue
			}
			for _, userId := range sendTo {
				fn_log.Printf("before publish cost = %v", time.Since(start))
				sendNum := c.publish(content, userId)
				fn_log.Printf("publish cost = %v", time.Since(start))
				if sendNum == 0 && holder.holder.Push {
					//push
				}
			}
		}
	}()
}
func (c *chatRoomService) asyncSaveMsg() {
	go func() {
		for holder := range c.saveChn {
			content, err := proto.Marshal(holder.payload)
			if err != nil {
				fn_log.Printf("Publish message -> marshal error  message Content = %v  error = %v ", content, err)
				continue
			}
			// 主要是 读扩散 还是写扩散  保存的方式不一样
			//读扩散基于chat     谢扩散基于个人
			if holder.holder.Save {
				if err := c.MqClient.PublishExchange(fc_constant.ImDbMessageSaveExchange, fc_constant.ImDbMessageSaveKey, content); err != nil {
					fn_log.Printf("%v", err)
				} else {
					//retry
				}
			}
		}
	}()
}

func (c *chatRoomService) publish(content []byte, userId int64) int64 {
	//todo 加一个chain 批次处理消息
	pubChannel := fmt.Sprintf(UserSubscibeChannelTemplate, userId)
	num, err := c.Pubsub.Publish(pubChannel, content)
	if err == nil {
		return num
	}
	fn_log.Printf("message publish %v", err)
	return 0
}

func (c *chatRoomService) PostInitilization() {
	c.asyncSaveMsg()
	c.asyncPubMsg()
	go func() {
		queue, err := c.MqClient.ListenExchangeWithQueue(fc_constant.ImDbMessageSaveExchange, fc_constant.ImDbMessageSaveQueue, fc_constant.ImDbMessageSaveKey)
		if err != nil {
			panic(err)
		}

		for msg := range queue {
			err = c.MqClient.Ack(msg.DeliveryTag)
			if err != nil {
				fn_log.Printf("ack error %v", err)
			}
			var message gatewaypb.MessagePayload
			err = proto.Unmarshal(msg.Body, &message)
			if err != nil {
				fn_log.Printf("message Unmarshal error %v", err)
				continue
			}
			err = c.MessageRepo.SaveMessage(message.ServerId, message.Sender, message.ChatId, msg.Body)
			if err != nil {
				fn_log.Printf("message save failed error %v", err)
			}
		}
	}()
}

func NewChatRoomControl() ChatRoomService {
	return &chatRoomService{
		saveChn: make(chan ActionMessage, 10),
		pubChn:  make(chan ActionMessage, 10),
	}
}
