package service_message

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	grpc_common "freechat/im/grpc-common"
	"github.com/liyanze888/funny-core/fn_log"
	"sync"
)

type ConnectHolder struct {
	wg *sync.WaitGroup
	//回头把输出做成异步 通过redis 接受接收数据
	Server      gatewaypb.ImService_ConnectServer
	worker      MessageDispatcher
	UserContext *grpc_common.GrpcContextInfo
}

func (c *ConnectHolder) Start() {
	c.wg.Add(1)
	defer c.wg.Done()
	for {
		message, err := c.Server.Recv()
		if err != nil {
			fn_log.Printf("%v", err)
			break
		}
		c.worker.Dispatch(message, c)
	}
}

func NewConnectHolder(wg *sync.WaitGroup, server gatewaypb.ImService_ConnectServer, worker MessageDispatcher) (*ConnectHolder, error) {
	// 初始化
	info, err := grpc_common.Tranfer2ContextInfo(server.Context())
	if err != nil {
		return nil, err
	}
	fn_log.Printf("%v", *info)
	return &ConnectHolder{
		wg:          wg,
		Server:      server,
		worker:      worker,
		UserContext: info,
	}, nil
}
