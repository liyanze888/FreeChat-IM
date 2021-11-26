package services

import (
	"freechat/im/config"
	"freechat/im/fc_constant"
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	serialpb "freechat/im/generated/grpc/im/server/message/serial"
	"freechat/im/rabbit"
	"freechat/im/repository"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewMessageServiceImpl())
}

type MessageService interface {
	readDiffusionSave(rmsg amqp.Delivery) error
	writeDiffusionSave(rmsg amqp.Delivery) error
}

type messageServiceImpl struct {
	MessageReadDiffusionRepo  repository.MessageReadDiffusionRepository  `autowire:""`
	MessageWriteDiffusionRepo repository.MessageWriteDiffusionRepository `autowire:""`
	MqClient                  rabbit.RabbitmqClient                      `autowire:""`
	GlobalConfig              *config.GlobalConfig                       `autowire:""`
	saveFuncs                 map[int]saveFunc
}

type saveFunc func(rmsg amqp.Delivery) error

func (m *messageServiceImpl) readDiffusionSave(rmsg amqp.Delivery) error {
	var message gatewaypb.MessagePayload
	err := proto.Unmarshal(rmsg.Body, &message)
	if err != nil {
		fn_log.Printf("message Unmarshal error %v", err)
		return err
	}
	err = m.MessageReadDiffusionRepo.SaveMessage(message.ServerId, message.Sender, message.ChatId, rmsg.Body)
	if err != nil {
		fn_log.Printf("message save failed error %v", err)
		return err
	}
	return nil
}

func (m *messageServiceImpl) writeDiffusionSave(rmsg amqp.Delivery) error {
	var message serialpb.SaveMessage
	err := proto.Unmarshal(rmsg.Body, &message)
	if err != nil {
		fn_log.Printf("message Unmarshal error %v", err)
		return err
	}

	var contentPaylod gatewaypb.MessagePayload
	err = proto.Unmarshal(message.Content, &contentPaylod)
	if err != nil {
		fn_log.Printf("message Unmarshal error %v", err)
		return err
	}

	dbPayloads := make(map[int64]*repository.SaveMessageDbPayload)
	for _, userId := range message.UserIds {
		dbIndex := userId / repository.SliceDbBaseLimit
		var dbPaylod *repository.SaveMessageDbPayload = nil
		var ok bool
		if dbPaylod, ok = dbPayloads[dbIndex]; !ok {
			dbPaylod = &repository.SaveMessageDbPayload{
				Payloads: make(map[int64][]*repository.SaveMessagePayload),
			}
			dbPayloads[dbIndex] = dbPaylod
		}

		tableIndex := userId / repository.SliceTableBaseLimit
		var tablePayload []*repository.SaveMessagePayload = nil
		if tablePayload, ok = dbPaylod.Payloads[tableIndex]; !ok {
			tablePayload = make([]*repository.SaveMessagePayload, 0)
			dbPaylod.Payloads[tableIndex] = tablePayload
		}

		dbPaylod.Payloads[tableIndex] = append(dbPaylod.Payloads[tableIndex], &repository.SaveMessagePayload{
			Id:      contentPaylod.ServerId,
			ChatId:  contentPaylod.ChatId,
			Owner:   userId,
			Sender:  contentPaylod.Sender,
			Content: message.Content,
		})
	}

	err = m.MessageWriteDiffusionRepo.SaveWriteDiffusionSave(dbPayloads)
	if err != nil {
		fn_log.Printf("message SaveWriteDiffusionSave error %v", err)
		return err
	}
	return nil
}

func (m *messageServiceImpl) PostInitilization() {
	m.saveFuncs[config.ReadDiffusion] = m.readDiffusionSave
	m.saveFuncs[config.WriteDiffusion] = m.writeDiffusionSave
	go func() {
		queue, err := m.MqClient.ListenExchangeWithQueue(fc_constant.ImDbMessageSaveExchange, fc_constant.ImDbMessageSaveQueue, fc_constant.ImDbMessageSaveKey)
		if err != nil {
			panic(err)
		}

		for msg := range queue {
			if err := m.saveFuncs[m.GlobalConfig.Diffusion](msg); err != nil {
				err = m.MqClient.UnAck(msg.DeliveryTag, true)
			} else {
				err = m.MqClient.Ack(msg.DeliveryTag)
				if err != nil {
					fn_log.Printf("ack error %v", err)
				}
			}

		}
	}()
}

func NewMessageServiceImpl() MessageService {
	return &messageServiceImpl{
		saveFuncs: make(map[int]saveFunc),
	}
}
