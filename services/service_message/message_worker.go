package service_message

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	grpc_common "freechat/im/grpc-common"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewMessageWorker())
}

type WorkMessageHolder struct {
	Payloads []*WorkMessagePaylaod
	SendTo   []int64 //len(sendTo) == 0 means Send All otherwise sendTo targets
	Save     bool    // is ζδΉε
	Push     bool
}

type WorkMessagePaylaod struct {
	Message *gatewaypb.MessageCell
}

type MessageTypeWorker interface {
	work(message *gatewaypb.MessageCell, user *grpc_common.GrpcContextInfo) (*WorkMessagePaylaod, error)
}

type MessageDispatcher interface {
	Dispatch(message *gatewaypb.MessagePayload, user *grpc_common.GrpcContextInfo) (*WorkMessageHolder, error)
}

type messageDispatcher struct {
	Workers map[int]MessageTypeWorker `autowire:"" beanFiledName:"workerType"`
}

func (m *messageDispatcher) Dispatch(payload *gatewaypb.MessagePayload, user *grpc_common.GrpcContextInfo) (*WorkMessageHolder, error) {
	fn_log.Printf("work payload = %v  info = %v", payload, *user)
	var payloads []*WorkMessagePaylaod
	for _, msg := range payload.GetItems() {
		finalMsg, err := m.Workers[int(msg.MsgType)].work(msg, user)
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, finalMsg)
	}

	return &WorkMessageHolder{
		Payloads: payloads,
		Push:     false,
		Save:     true,
	}, nil
}

func NewMessageWorker() MessageDispatcher {
	return &messageDispatcher{}
}
