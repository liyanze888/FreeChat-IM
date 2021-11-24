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

const (
	DefaultCombineMsgCount  = 5    //5调数据
	DefaultSendoutTimeAfter = 1500 //1.5s
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
	// 是否需要拼装数据
	// 拼装成完整的   是否需要combine data
	c.combineMessage = append(c.combineMessage, &message)
	if len(c.combineMessage) >= c.combineMsgLimit {
		c.sendOut()
	}
	return nil
}

func (c *UserStream) asyncTimeoutSendMsg() {
	go func() {
		chn := make(chan struct{})
		defer close(chn)
		for c.timeLoop {
			c.afterFewSecond(chn)
			select {
			case <-chn:
				c.sendOut()
			}
		}
		fn_log.Printf("asyncTimeoutSendMsg finish")
	}()
}

func (c *UserStream) finish() error {
	c.timeLoop = false
	return nil
}

func (c *UserStream) afterFewSecond(chn chan struct{}) {
	time.AfterFunc(time.Duration(c.timeAfterSend)*time.Millisecond, func() {
		chn <- struct{}{}
	})
}

func (c *UserStream) sendOut() {
	c.combineMutex.Lock()
	defer c.combineMutex.Unlock()
	if len(c.combineMessage) > 0 {
		_ = c.server.Send(
			&gatewaypb.MessageWrapper{
				Payloads: c.combineMessage,
				Count:    int64(len(c.combineMessage)),
			})
		c.combineMessage = c.combineMessage[0:0]
	}
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
	if err = u.Pubsub.SubcribeChannel(server.Context(), fmt.Sprintf(UserSubscibeChannelTemplate, user.UserId), stream.send, stream.finish); err != nil {
		return err
	}
	stream.Start(user)
	wg.Wait()
	return nil
}

func (u *UserStreamFactory) startNewUserStream(wg *sync.WaitGroup, server gatewaypb.ImService_ConnectServer) *UserStream {
	userStream := &UserStream{
		wg:              wg,
		chatRoomControl: u.ChatRoomControl,
		server:          server,
		timeLoop:        true,
		combineMsgLimit: DefaultCombineMsgCount,  //可配置
		timeAfterSend:   DefaultSendoutTimeAfter, //可配置
	}
	userStream.asyncTimeoutSendMsg()
	return userStream
}

// UserStream 接收消息 与发消息
type UserStream struct {
	wg              *sync.WaitGroup
	server          gatewaypb.ImService_ConnectServer
	chatRoomControl ChatRoomService
	combineMutex    sync.Mutex
	combineMessage  []*gatewaypb.MessagePayload
	timeLoop        bool
	combineMsgLimit int   //限制长度
	timeAfterSend   int64 //多长时间后自动发送 unit ms
}

type UserStreamFactory struct {
	Worker          MessageDispatcher        `autowire:""`
	ChatRoomControl ChatRoomService          `autowire:""`
	Pubsub          *subscribe.PubSubManager `autowire:""`
}

func NewUserStreamFactory() *UserStreamFactory {
	return &UserStreamFactory{}
}
