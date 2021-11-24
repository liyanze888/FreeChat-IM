package subscribe

//
//import (
//	"context"
//	"fmt"
//	"freechat/im/fc_utils/str_utils"
//	gatewaypb "freechat/im/generated/grpc/im/gateway"
//	grpcCommon "freechat/im/grpc-common"
//	"github.com/go-redis/redis/v8"
//	"github.com/liyanze888/funny-core/fn_factory"
//	"github.com/liyanze888/funny-core/fn_log"
//	"google.golang.org/protobuf/proto"
//	"reflect"
//	"sync"
//)
//
//const (
//	UserSubscibeChannelTemplate = "message:subscribe:%d"
//)
//
//type Unsubscribe func()
//
//func init() {
//	fn_factory.BeanFactory.RegisterBean(NewUserContextFactory())
//}
//
//type UserContext struct {
//	Rdb        redis.UniversalClient `autowire:""`
//	UserInfo   *grpcCommon.GrpcContextInfo
//	subChannel string
//	factory    UserContextFactory
//	Server     gatewaypb.ImService_ConnectServer
//	msgLoop    bool
//}
//
//func (m *UserContext) StartUpListen() {
//	m.startListenSendMessage()
//}
//
//func (m *UserContext) startListenSendMessage() {
//	var wg sync.WaitGroup
//	wg.Add(1)
//	go func() {
//		release := false
//
//		defer func() {
//			if !release {
//				wg.Done()
//			}
//		}()
//		wg.Done()
//		release = true
//		ctx := context.Background()
//		for m.msgLoop {
//			receive, err := m.subscriber.ReceiveMessage(ctx)
//			fn_log.Printf("receive")
//			if err != nil {
//				fn_log.Printf("%s reciveMessage error %v %v", m.subChannel, err, reflect.TypeOf(err).Elem().Name())
//				continue
//			}
//			var message gatewaypb.MessagePayload
//			err = proto.Unmarshal(str_utils.Str2Bytes(receive.Payload), &message)
//			if err != nil {
//				fn_log.Printf("%s proto.Unmarshal res = %v error %v", m.subChannel, receive.Payload, err)
//				continue
//			}
//			// 拼装成完整的   是否需要combine data
//			m.Server.Send(
//				&gatewaypb.MessageWrapper{
//					Payloads: []*gatewaypb.MessagePayload{&message}})
//		}
//		fn_log.Printf("finish")
//	}()
//	wg.Wait()
//}
//
//func (m *UserContext) Unsubscribe() {
//	m.msgLoop = false
//	err := m.subscriber.Unsubscribe(context.Background(), m.subChannel)
//	if err != nil {
//		fn_log.Printf("%v", err)
//	}
//	err = m.subscriber.Close()
//	if err != nil {
//		fn_log.Printf("%v", err)
//	}
//}
//
//type UserContextFactory struct {
//	Rdb         redis.UniversalClient `autowire:""`
//	pubsub      *redis.PubSub
//	subscribers map[string]chan *redis.Message
//}
//
//func (mf *UserContextFactory) PostInitilization() {
//	mf.pubsub = mf.Rdb.Subscribe(context.Background())
//}
//
//func (mf UserContextFactory) NewUserContext(info *grpcCommon.GrpcContextInfo, server gatewaypb.ImService_ConnectServer) (*UserContext, error) {
//	channel := fmt.Sprintf(UserSubscibeChannelTemplate, info.UserId)
//	consumer := &UserContext{
//		UserInfo:   info,
//		Rdb:        mf.Rdb,
//		Server:     server,
//		msgLoop:    true,
//		factory:    mf,
//		subChannel: channel,
//	}
//	err := mf.pubsub.Subscribe(context.Background(), channel)
//	if err != nil {
//		fn_log.Printf("mf.pubsub.Subscribe  = %v", err)
//		return nil, err
//	}
//	msgChn := make(chan *redis.Message)
//	consumer.StartUpListen()
//	return consumer, nil
//}
//
//// 订阅工厂
//func NewUserContextFactory() *UserContextFactory {
//	return &UserContextFactory{
//		subscribers: make(map[string]chan *redis.Message),
//	}
//}
