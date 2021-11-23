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
	fn_factory.BeanFactory.RegisterBean(NewUserContextFactory())
}

type UserContext struct {
	Rdb        redis.UniversalClient `autowire:""`
	UserInfo   *grpcCommon.GrpcContextInfo
	subChannel string
	subscriber *redis.PubSub
	Server     gatewaypb.ImService_ConnectServer
}

func (m *UserContext) StartUpListen() {
	m.startListenSendMessage()
}

func (m *UserContext) startListenSendMessage() {
	go func() {
		m.subChannel = fmt.Sprintf(UserSubscibeChannelTemplate, m.UserInfo.UserId)
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
			m.Server.Send(&message)
		}
	}()
}

func (m *UserContext) Publish(message *gatewaypb.MessageWrapper) {
	//todo 加一个chain 批次处理消息
	marshal, err := proto.Marshal(message)
	if err != nil {
		fn_log.Printf("Publish message -> marshal error  message Content = %v  error = %v ", message, err)
	}
	m.Rdb.Publish(context.Background(), m.subChannel, marshal)
}

func (m *UserContext) Unsubscribe() {
	m.subscriber.Unsubscribe(context.Background(), m.subChannel)
}

type UserContextFactory struct {
	Rdb redis.UniversalClient `autowire:""`
}

func (mf UserContextFactory) NewUserContext(info *grpcCommon.GrpcContextInfo, server gatewaypb.ImService_ConnectServer) *UserContext {
	consumer := &UserContext{
		UserInfo: info,
		Rdb:      mf.Rdb,
		Server:   server,
	}
	consumer.StartUpListen()
	return consumer
}

// 订阅工厂
func NewUserContextFactory() *UserContextFactory {
	return &UserContextFactory{}
}
