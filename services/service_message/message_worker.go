package service_message

import (
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	"freechat/im/repository"
	"freechat/im/subscribe"
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
	work(message *gatewaypb.MessageCell, user *subscribe.UserContext) (*MessageHolder, error)
}

type MessageDispatcher interface {
	Dispatch(message *gatewaypb.MessageWrapper, user *subscribe.UserContext) (*gatewaypb.MessageWrapper, error)
}

type messageDispatcher struct {
	Workers     map[int]MessageTypeWorker    `autowire:"" beanFiledName:"workerType"`
	MessageRepo repository.MessageRepository `autowire:""`
	IdUtils     fn_utils.IdUtils             `autowire:""`
}

func (m *messageDispatcher) Dispatch(message *gatewaypb.MessageWrapper, user *subscribe.UserContext) (*gatewaypb.MessageWrapper, error) {
	fn_log.Printf("work msg = %v  info = %v", message, *user)
	for _, msg := range message.GetMessage().GetItems() {
		chatId := msg.GetChatId()
		hodler, err := m.Workers[int(msg.MsgType)].work(msg, user)
		if err != nil {
			return nil, err
		}
		id, err := m.IdUtils.NextID()
		if err != nil {
			return nil, err
		}

		content, err := proto.Marshal(hodler.Message)
		if err != nil {
			return nil, err
		}
		fn_log.Printf("%v", hodler.Message)
		//todo 前置处理消息器
		go func() {
			//TODO 发送到mq 中 异步消费  如果以后做集群 则做主从同步 然后做cdn 分发
			err = m.MessageRepo.SaveMessage(int64(id), user.UserInfo.UserId, chatId, content)
			if err != nil {
				fn_log.Printf(err.Error())
			}
		}()
	}
	return message, nil
}

func NewMessageWorker() MessageDispatcher {
	return &messageDispatcher{}
}
