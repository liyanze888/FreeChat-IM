package service_message

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	"freechat/im/repository"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
	"github.com/liyanze888/funny-core/fn_utils"
	"google.golang.org/protobuf/proto"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewMessageWorker())
}

type MessageHolder struct {
	Message *gatewaypb.MessageCell
	SendTo  []int64 //len(sendTo) == 0 means Send All otherwise sendTo targets
	Save    bool    // is 持久化
	Push    bool
}

type MessageTypeWorker interface {
	work(message *gatewaypb.MessageCell, info *ConnectHolder) (*MessageHolder, error)
}

type MessageDispatcher interface {
	Dispatch(message *gatewaypb.MessageWrapper, info *ConnectHolder) error
}

type messageDispatcher struct {
	Workers     map[int]MessageTypeWorker    `autowire:"" beanFiledName:"workerType"`
	MessageRepo repository.MessageRepository `autowire:""`
	IdUtils     fn_utils.IdUtils             `autowire:""`
}

func (m *messageDispatcher) Dispatch(message *gatewaypb.MessageWrapper, info *ConnectHolder) error {
	fn_log.Printf("work msg = %v  info = %v", message, *info)
	for _, msg := range message.GetMessage().GetItems() {
		hodler, err := m.Workers[int(msg.MsgType)].work(msg, info)
		if err != nil {
			return err
		}
		id, err := m.IdUtils.NextID()
		if err != nil {
			return err
		}

		content, err := proto.Marshal(hodler.Message)
		if err != nil {
			return err
		}
		fn_log.Printf("%v", hodler.Message)
		err = m.MessageRepo.SaveMessage(int64(id), info.UserContext.UserId, msg.GetChatId(), content)
		if err != nil {
			fn_log.Printf(err.Error())
		}
	}
	return nil
}

func NewMessageWorker() MessageDispatcher {
	return &messageDispatcher{}
}
