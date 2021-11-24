package service_message

import (
	"fmt"
	"freechat/im/fc_utils/str_utils"
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	grpc_common "freechat/im/grpc-common"
	"freechat/im/subscribe"
	"github.com/go-redis/redis/v8"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewUserStreamFactory())
}

const (
	UserSubscibeChannelTemplate = "message:subscribe:%d"
)

func (c *UserStream) Start(user *grpc_common.GrpcContextInfo) {
	c.wg.Add(1)
	defer c.wg.Done()
	for {
		message, err := c.server.Recv()
		if err != nil {
			fn_log.Printf("%v", err)
			break
		}
		start := time.Now()
		c.chatRoomControl.SendMsg(message, user)
		fn_log.Printf("cost =  %v", time.Since(start))
	}
}

func (c *UserStream) send(receive *redis.Message) error {
	var message gatewaypb.MessagePayload
	err := proto.Unmarshal(str_utils.Str2Bytes(receive.Payload), &message)
	if err != nil {
		fn_log.Printf("%s proto.Unmarshal res = %v error %v", receive.Channel, receive.Payload, err)
	}
	// 拼装成完整的   是否需要combine data
	return c.server.Send(
		&gatewaypb.MessageWrapper{
			Payloads: []*gatewaypb.MessagePayload{&message}})
}

func (u *UserStreamFactory) StartNewUserStream(server gatewaypb.ImService_ConnectServer) error {
	// 初始化
	var wg sync.WaitGroup
	user, err := grpc_common.Tranfer2ContextInfo(server.Context())
	if err != nil {
		return err
	}
	fn_log.Printf("%v", *user)
	stream := u.startNewUserStream(&wg, server)
	if err = u.Pubsub.SubcribeChannel(server.Context(), fmt.Sprintf(UserSubscibeChannelTemplate, user.UserId), stream.send); err != nil {
		return err
	}
	stream.Start(user)
	wg.Wait()
	return nil
}

func (u *UserStreamFactory) startNewUserStream(wg *sync.WaitGroup, server gatewaypb.ImService_ConnectServer) *UserStream {
	userHodler := &UserStream{
		wg:              wg,
		chatRoomControl: u.ChatRoomControl,
		server:          server,
	}
	return userHodler
}

// UserStream 接收消息 与发消息
type UserStream struct {
	wg              *sync.WaitGroup
	server          gatewaypb.ImService_ConnectServer
	chatRoomControl ChatRoomService
}

type UserStreamFactory struct {
	Worker          MessageDispatcher        `autowire:""`
	ChatRoomControl ChatRoomService          `autowire:""`
	Pubsub          *subscribe.PubSubManager `autowire:""`
}

func NewUserStreamFactory() *UserStreamFactory {
	return &UserStreamFactory{}
}
