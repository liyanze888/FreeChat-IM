package service_message

import (
	"fmt"
	"freechat/im/config"
	"freechat/im/fc_constant"
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	serialpb "freechat/im/generated/grpc/im/server/message/serial"
	grpc_common "freechat/im/grpc-common"
	"freechat/im/rabbit"
	"freechat/im/subscribe"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
	"google.golang.org/protobuf/proto"
	"time"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewAsyncActionMessage())
}

type ActionMessage struct {
	sendTo   []int64
	Save     bool // is 持久化
	Push     bool
	payload  *gatewaypb.MessagePayload
	userInfo *grpc_common.GrpcContextInfo
}

type AsyncActionMessage interface {
	async(am ActionMessage)
}

type asyncActionMessage struct {
	GlobalConfig *config.GlobalConfig     `autowire:""`
	Pubsub       *subscribe.PubSubManager `autowire:""`
	MqClient     rabbit.RabbitmqClient    `autowire:""`
	saveChn      chan ActionMessage
	pubChn       chan ActionMessage
}

func (a *asyncActionMessage) async(am ActionMessage) {
	go func() {
		if len(am.sendTo) == 0 {
			// get all by conversation id
			am.sendTo = []int64{am.userInfo.UserId}
		}
		a.pubChn <- am
		a.saveChn <- am
	}()
}

func (a *asyncActionMessage) asyncPubMsg() {
	go func() {
		for holder := range a.pubChn {
			start := time.Now()
			sendTo := holder.sendTo
			content, err := proto.Marshal(holder.payload)
			if err != nil {
				fn_log.Printf("Publish message -> marshal error  message Content = %v  error = %v ", content, err)
				continue
			}
			for _, userId := range sendTo {
				fn_log.Printf("before publish cost = %v", time.Since(start))
				sendNum := a.publish(content, userId)
				fn_log.Printf("publish cost = %v", time.Since(start))
				if sendNum == 0 && holder.Push {
					//push
				}
			}
		}
	}()
}

func (a *asyncActionMessage) publish(content []byte, userId int64) int64 {
	//todo 加一个chain 批次处理消息
	pubChannel := fmt.Sprintf(UserSubscibeChannelTemplate, userId)
	num, err := a.Pubsub.Publish(pubChannel, content)
	if err == nil {
		return num
	}
	fn_log.Printf("message publish %v", err)
	return 0
}

func (a *asyncActionMessage) asyncSaveMsg() {
	go func() {
		for holder := range a.saveChn {
			content, err := proto.Marshal(holder.payload)
			if err != nil {
				fn_log.Printf("Publish message -> marshal error  message Content = %v  error = %v ", content, err)
				continue
			}
			if a.GlobalConfig.Diffusion == config.ReadDiffusion {
				// 主要是 读扩散 还是写扩散  保存的方式不一样
				//读扩散基于chat     谢扩散基于个人
				if holder.Save {
					if err := a.MqClient.PublishExchange(fc_constant.ImDbMessageSaveExchange, fc_constant.ImDbMessageSaveKey, content); err != nil {
						fn_log.Printf("%v", err)
					} else {
						//retry
					}
				}
			} else {
				if holder.Save {
					msgPaylod := serialpb.SaveMessage{
						Content: content,
						UserIds: holder.sendTo,
					}
					content, err := proto.Marshal(&msgPaylod)
					if err != nil {
						fn_log.Printf("Publish message -> marshal error  message Content = %v  error = %v ", content, err)
						continue
					}
					if err := a.MqClient.PublishExchange(fc_constant.ImDbMessageSaveExchange, fc_constant.ImDbMessageSaveKey, content); err != nil {
						fn_log.Printf("%v", err)
					}
				}
			}
		}
	}()
}

func (a *asyncActionMessage) PostInitilization() {
	a.asyncSaveMsg()
	a.asyncPubMsg()
}

func NewAsyncActionMessage() AsyncActionMessage {
	return &asyncActionMessage{
		saveChn: make(chan ActionMessage, 10),
		pubChn:  make(chan ActionMessage, 10),
	}
}
