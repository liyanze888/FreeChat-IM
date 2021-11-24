package subscribe

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
	"runtime/debug"
	"sync"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewPubSubManager())
}

type OnMessage func(*redis.Message) error
type OnFinish func() error

type PubSubManager struct {
	Rdb         redis.UniversalClient `autowire:""`
	pubsub      *redis.PubSub
	subChannels map[string]*ChannelScuscribersManager
	psMutex     sync.Mutex
}

// 同一channel 管理器
type ChannelScuscribersManager struct {
	csMutex       sync.Mutex
	chnIdChannels map[int64]chan *redis.Message
	chId          int64
}

func (p *PubSubManager) Publish(channel string, content []byte) (int64, error) {
	fn_log.Printf("Publish %v", channel)
	publish := p.Rdb.Publish(context.Background(), channel, content)
	if publish.Err() != nil {
		fn_log.Printf("Publish %v", publish.Err())
		return 0, publish.Err()
	}
	return publish.Val(), nil
}

func (p *PubSubManager) SubcribeChannel(ctx context.Context, channel string, f OnMessage, c OnFinish) error {
	p.psMutex.Lock()
	defer p.psMutex.Unlock()
	if _, ok := p.subChannels[channel]; !ok {
		if err := p.pubsub.Subscribe(context.Background(), channel); err != nil {
			return err
		}
		p.subChannels[channel] = &ChannelScuscribersManager{
			chnIdChannels: make(map[int64]chan *redis.Message),
		}
	}
	fn_log.Printf("%v", len(p.subChannels))
	return p.addCh(ctx, channel, f, c)
}

func (p *PubSubManager) addCh(ctx context.Context, channel string, f OnMessage, c OnFinish) error {
	manager := p.subChannels[channel]
	manager.csMutex.Lock()
	defer manager.csMutex.Unlock()
	manager.chId++
	chId := manager.chId
	msgChn := make(chan *redis.Message)
	manager.chnIdChannels[chId] = msgChn
	fn_log.Printf("%v", len(manager.chnIdChannels))
	go func() {
		loop := true
		defer close(msgChn)
		for loop {
			select {
			case msg := <-msgChn:
				Call(f, msg)
			case <-ctx.Done():
				fn_log.Printf("%v")
				loop = false
				c()
				p.rmCh(channel, chId)
			}
		}
	}()
	return nil
}

func Call(f OnMessage, msg *redis.Message) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			fn_log.Printf("%s\n", err)
		}
	}()
	if f != nil {
		err := f(msg)
		fn_log.Printf("%s\n", err)
	}
}

func (p *PubSubManager) rmCh(channel string, chId int64) {
	p.psMutex.Lock()
	defer p.psMutex.Unlock()
	manager := p.subChannels[channel]
	manager.csMutex.Lock()
	defer manager.csMutex.Unlock()
	delete(manager.chnIdChannels, chId)
	if len(manager.chnIdChannels) == 0 {
		delete(p.subChannels, channel)
		if err := p.pubsub.Unsubscribe(context.Background(), channel); err != nil {
			fn_log.Printf("Unsubscribe  %v", err)
		}
	}
}

func (p *PubSubManager) startListen() {
	go func() {
		for {
			message, err := p.pubsub.ReceiveMessage(context.Background())
			if err != nil {
				fn_log.Printf("ReceiveMessage %v", err)
			}

			if sp, ok := p.subChannels[message.Channel]; ok {
				for _, val := range sp.chnIdChannels {
					val <- message
				}
			}
		}
	}()
}

func (p *PubSubManager) PostInitilization() {
	p.pubsub = p.Rdb.Subscribe(context.Background())
	p.startListen()
}

func NewPubSubManager() *PubSubManager {
	return &PubSubManager{
		subChannels: make(map[string]*ChannelScuscribersManager),
	}
}
