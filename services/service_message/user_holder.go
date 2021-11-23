package service_message

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	grpc_common "freechat/im/grpc-common"
	"freechat/im/subscribe"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
	"sync"
	"time"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewUserHodlerFactory())
}

func (c *UserStream) Start() {
	c.wg.Add(1)
	defer c.wg.Done()
	for {
		message, err := c.UserContext.Server.Recv()
		start := time.Now()
		if err != nil {
			fn_log.Printf("%v", err)
			c.UserContext.Unsubscribe()
			break
		}
		c.chatRoomControl.SendMsg(message, c.UserContext)
		fn_log.Printf("cost =  %v", time.Since(start))
	}
}

func (u *UserStreamFactory) StartNewUserStream(server gatewaypb.ImService_ConnectServer) error {
	// 初始化
	var wg sync.WaitGroup
	info, err := grpc_common.Tranfer2ContextInfo(server)
	if err != nil {
		return err
	}
	fn_log.Printf("%v", *info)
	userHolder := u.startNewUserStream(info, &wg, server)
	userHolder.Start()
	wg.Wait()
	return nil
}

type UserStream struct {
	wg              *sync.WaitGroup
	UserContext     *subscribe.UserContext
	msFactory       *subscribe.UserContextFactory
	chatRoomControl ChatRoomService
}

type UserStreamFactory struct {
	Worker          MessageDispatcher             `autowire:""`
	MsFactory       *subscribe.UserContextFactory `autowire:""`
	ChatRoomControl ChatRoomService               `autowire:""`
}

func NewUserHodlerFactory() *UserStreamFactory {
	return &UserStreamFactory{}
}

func (u *UserStreamFactory) startNewUserStream(userInfo *grpc_common.GrpcContextInfo, wg *sync.WaitGroup, server gatewaypb.ImService_ConnectServer) *UserStream {
	userContext := u.MsFactory.NewUserContext(userInfo, server)
	userHodler := &UserStream{
		wg:              wg,
		chatRoomControl: u.ChatRoomControl,
		msFactory:       u.MsFactory,
		UserContext:     userContext,
	}
	return userHodler
}
