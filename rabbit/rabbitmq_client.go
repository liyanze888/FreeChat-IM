package rabbit

import (
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
	"github.com/streadway/amqp"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewRabbitmqClient())
}

type RabbitmqClient interface {
	PublishQueue(keyName string, data []byte) error
	PublishExchange(exchangeName, keyName string, data []byte) error
	ListenQueue(queueName string) (<-chan amqp.Delivery, error)
	ListenExchangeWithQueue(exchangeName, queueName, keyName string) (<-chan amqp.Delivery, error)
	Ack(id uint64) error
	UnAck(id uint64, requeue bool) error
}

type rabbitmq struct {
	conn     *amqp.Connection
	queueMap map[string]<-chan amqp.Delivery
	channel  *amqp.Channel
}

func (r *rabbitmq) PublishQueue(keyName string, data []byte) error {
	err := r.channel.Publish("", keyName, false, false, amqp.Publishing{
		Body: data,
	})
	return err
}

func (r *rabbitmq) PublishExchange(exchangeName, keyName string, data []byte) error {
	err := r.channel.Publish(exchangeName, keyName, false, false, amqp.Publishing{
		Body: data,
	})
	return err
}

func (r *rabbitmq) ListenQueue(queueName string) (<-chan amqp.Delivery, error) {

	if value, ok := r.queueMap[queueName]; ok {
		return value, nil
	}

	_, err := r.channel.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		panic(err)
	}
	revicer, err2 := r.channel.Consume(queueName, "", false, false, false, false, nil)
	if err2 != nil {
		panic(err2)
	}
	r.queueMap[queueName] = revicer
	return revicer, nil
}

func (r *rabbitmq) ListenExchangeWithQueue(exchangeName, queueName, keyName string) (<-chan amqp.Delivery, error) {
	if value, ok := r.queueMap[queueName]; ok {
		return value, nil
	}

	err := r.channel.ExchangeDeclare(exchangeName, "topic", false, true, false, false, nil)
	if err != nil {
		panic(err)
	}
	_, err = r.channel.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		panic(err)
	}

	err = r.channel.QueueBind(queueName, keyName, exchangeName, false, nil)
	if err != nil {
		panic(err)
	}

	revicer, err2 := r.channel.Consume(queueName, "", false, false, false, false, nil)
	if err2 != nil {
		panic(err2)
	}
	r.queueMap[queueName] = revicer

	return revicer, nil
}

func (r *rabbitmq) Ack(id uint64) error {
	return r.channel.Ack(id, false)
}

func (r *rabbitmq) UnAck(id uint64, requeue bool) error {
	//或者false 本地自己记录
	return r.channel.Nack(id, false, requeue)
}

func NewRabbitmqClient() RabbitmqClient {
	url := "amqp://admin:admin@152.136.28.100:5672/"
	conn, err := amqp.Dial(url)
	if err != nil {
		fn_log.Printf("rabbitmq %v", err)
		panic(err)
	}
	channel, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	return &rabbitmq{
		conn:     conn,
		channel:  channel,
		queueMap: make(map[string]<-chan amqp.Delivery),
	}
}
