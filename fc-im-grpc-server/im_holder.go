package fc_im_grpc_server

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	"github.com/liyanze888/funny-core/fn_log"
	"sync"
)

type ConnectHolder struct {
	wg *sync.WaitGroup
	//回头把输出做成异步 通过redis 接受接收数据
	Server gatewaypb.ImService_ConnectServer
	worker MessageWorker
}

func (c *ConnectHolder) Start() {
	c.wg.Add(1)
	defer c.wg.Done()
	for {
		recv, err := c.Server.Recv()
		if err != nil {
			fn_log.Printf("%v", err)
			break
		}
		c.worker.Work(recv, c)
	}
}

func NewConnectHolder(wg *sync.WaitGroup, server gatewaypb.ImService_ConnectServer, worker MessageWorker) *ConnectHolder {
	return &ConnectHolder{
		wg:     wg,
		Server: server,
		worker: worker,
	}
}
