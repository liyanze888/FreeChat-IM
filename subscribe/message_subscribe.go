package subscribe

import (
	"context"
	"fmt"
	"freechat/im/fc_utils/str_utils"
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	grpcCommon "freechat/im/grpc-common"
	"github.com/go-redis/redis/v8"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
	"google.golang.org/protobuf/proto"
)

const (
	UserSubscibeChannelTemplate = "message:subscribe:%d"
)

type Unsubscribe func()

func init() {
	fn_factory.BeanFactory.RegisterBean(NewMessageSubscribeFactory())
}

type MessageSubscribeConsumer struct {
	Rdb        redis.UniversalClient `autowire:""`
	info       *grpcCommon.GrpcContextInfo
	subChannel string
	subscriber *redis.PubSub
}

func (m *MessageSubscribeConsumer) StartUpListen() {
	go func() {
		m.subChannel = fmt.Sprintf(UserSubscibeChannelTemplate, m.info.UserId)
		m.subscriber = m.Rdb.Subscribe(context.Background(), m.subChannel)
		ctx := context.Background()
		for {
			receive, err := m.subscriber.ReceiveMessage(ctx)
			fn_log.Printf("receive")
			if err != nil {
				fn_log.Printf("%s reciveMessage error %v", m.subChannel, err)
				continue
			}
			var message gatewaypb.MessageWrapper
			err = proto.Unmarshal(str_utils.Str2Bytes(receive.Payload), &message)
			if err != nil {
				fn_log.Printf("%s proto.Unmarshal res = %v error %v", m.subChannel, receive.Payload, err)
				continue
			}
			m.info.Server.Send(&message)
		}
	}()
}

func (m *MessageSubscribeConsumer) Publish(message *gatewaypb.MessageWrapper) {
	marshal, err := proto.Marshal(message)
	if err != nil {
		fn_log.Printf("Publish message -> marshal error  message Content = %v  error = %v ", message, err)
	}
	m.Rdb.Publish(context.Background(), m.subChannel, marshal)
}

func (m *MessageSubscribeConsumer) Unsubscribe() {
	m.subscriber.Unsubscribe(context.Background(), m.subChannel)
}

type MessageSubscribeFactory struct {
	Rdb redis.UniversalClient `autowire:""`
}

func (mf MessageSubscribeFactory) CreateMessageSubscribeConsumer(info *grpcCommon.GrpcContextInfo) *MessageSubscribeConsumer {
	consumer := &MessageSubscribeConsumer{
		info: info,
		Rdb:  mf.Rdb,
	}
	consumer.StartUpListen()
	return consumer
}

// 订阅工厂
func NewMessageSubscribeFactory() *MessageSubscribeFactory {
	return &MessageSubscribeFactory{}
}
