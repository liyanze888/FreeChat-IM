package repository

import (
	. "freechat/im/generated/dsl"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/lqs/sqlingo"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewMessageRepository())
}

type MessageRepository interface {
	SaveMessage(id int64, userId int64, chatId int64, content []byte) error
	ListMessage(chatId int64) (models []*MessageModel, err error)
}

func (m messageRepository) SaveMessage(id int64, userId int64, chatId int64, content []byte) error {
	_, err := m.Db.InsertInto(Message).
		Fields(Message.Id, Message.UserId, Message.ChatId, Message.Content).
		Values(id, userId, chatId, string(content)).
		Execute()
	return err
}

type messageRepository struct {
	Db sqlingo.Database `autowire:""`
}

func (m messageRepository) ListMessage(chatId int64) (models []*MessageModel, err error) {
	_, err = m.Db.SelectFrom(Message).Where(Message.ChatId.Equals(chatId)).FetchAll(&models)
	return
}

func NewMessageRepository() MessageRepository {
	return &messageRepository{}
}
