package service_message

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	"freechat/im/repository"
	"freechat/im/subscribe"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewMessageWorker())
}

type WorkMessageHolder struct {
	Payloads []*WorkMessagePaylaod
	SendTo   []int64 //len(sendTo) == 0 means Send All otherwise sendTo targets
	Save     bool    // is 持久化
	Push     bool
}

type WorkMessagePaylaod struct {
	Message *gatewaypb.MessageCell
}

type MessageTypeWorker interface {
	work(message *gatewaypb.MessageCell, user *subscribe.UserContext) (*WorkMessagePaylaod, error)
}

type MessageDispatcher interface {
	Dispatch(message *gatewaypb.MessagePayload, user *subscribe.UserContext) (*WorkMessageHolder, error)
}

type messageDispatcher struct {
	Workers     map[int]MessageTypeWorker    `autowire:"" beanFiledName:"workerType"`
	MessageRepo repository.MessageRepository `autowire:""`
}

func (m *messageDispatcher) Dispatch(payload *gatewaypb.MessagePayload, user *subscribe.UserContext) (*WorkMessageHolder, error) {
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
