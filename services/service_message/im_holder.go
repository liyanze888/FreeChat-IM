package service_message

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	grpc_common "freechat/im/grpc-common"
	"freechat/im/subscribe"
	"github.com/liyanze888/funny-core/fn_log"
	"sync"
	"time"
)

type ConnectHolder struct {
	wg *sync.WaitGroup
	//回头把输出做成异步 通过redis 接受接收数据
	//Server      gatewaypb.ImService_ConnectServer
	worker      MessageDispatcher
	UserContext *grpc_common.GrpcContextInfo
	msFactory   *subscribe.MessageSubscribeFactory
}

func (c *ConnectHolder) Start() {
	c.wg.Add(1)
	defer c.wg.Done()
	Server := c.UserContext.Server
	consumer := c.msFactory.CreateMessageSubscribeConsumer(c.UserContext)
	consumer.StartUpListen()
	for {
		message, err := Server.Recv()
		start := time.Now()
		if err != nil {
			fn_log.Printf("%v", err)
			consumer.Unsubscribe()
			break
		}
		message, err = c.worker.Dispatch(message, c)
		if err != nil {
			fn_log.Printf("dispatch message error = %v  %v", message, err)
			// 错误处理
			continue
		}
		if message != nil {
			consumer.Publish(message)
		}
		fn_log.Printf("cost =  %v", time.Since(start))
	}
}

func NewConnectHolder(wg *sync.WaitGroup, server gatewaypb.ImService_ConnectServer, worker MessageDispatcher, msFactory *subscribe.MessageSubscribeFactory) (*ConnectHolder, error) {
	// 初始化
	info, err := grpc_common.Tranfer2ContextInfo(server)
	if err != nil {
		return nil, err
	}
	fn_log.Printf("%v", *info)
	return &ConnectHolder{
		wg:          wg,
		worker:      worker,
		UserContext: info,
		msFactory:   msFactory,
	}, nil
}
