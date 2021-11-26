package repository

import (
	"fmt"
	. "freechat/im/generated/dsl"
	"github.com/go-sql-driver/mysql"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_log"
	"github.com/lqs/sqlingo"
	"time"
)

func init() {
	fn_factory.BeanFactory.RegisterBean(NewMessageWriteDiffusionRepository())
}

const (
	SliceDbBaseLimit    = 10_000_000 //1kw 分DB
	SliceTableBaseLimit = 100_000    // 10w 分table
)

type MessageWriteDiffusionRepository interface {
	SaveWriteDiffusionSave(dbPlayIds map[int64]*SaveMessageDbPayload) error
	ListMessage(chatId int64, userId int64, lastTimestamp int64) (models []*MessageModel, err error)
}

const (
	CreateTableSql  = "CREATE TABLE message_%d(\n  id BIGINT NOT NULL ,\n  sender_id BIGINT NOT NULL COMMENT 'sender',\n  user_id BIGINT NOT NULL COMMENT 'owner',\n  chat_id BIGINT NOT NULL COMMENT 'conversation id',\n  content BLOB NOT NULL COMMENT 'message content' ,\n  create_at TIMESTAMP DEFAULT current_timestamp(),\n  PRIMARY KEY (`id`),\n  INDEX m_user_id (`user_id`),\n  INDEX m_chat_id_user_id (`chat_id` , `user_id`)\n)"
	InsertMsgSql    = "INSERT INTO message_%d(id , sender_id , user_id , chat_id  , content , create_at) Values %s"
	InsertMsgValSql = "(%d , %d , %d , %d, %v , %d)"
)

func (m *messageWriteDiffusionRepository) SaveWriteDiffusionSave(dbPlayIds map[int64]*SaveMessageDbPayload) error {
	if len(dbPlayIds) == 0 {
		return nil
	}

	fn_log.Printf("%v", dbPlayIds)

	for dbIndex, dbpayload := range dbPlayIds {
		for tableIndex, datas := range dbpayload.Payloads {
			dataSql := ""
			for _, dataPayload := range datas {
				if len(dataSql) > 0 {
					dataSql += ", " + fmt.Sprintf(InsertMsgValSql, dataPayload.Id, dataPayload.Sender, dataPayload.Owner, dataPayload.ChatId, dataPayload.Content, time.Now().UnixMilli())
				} else {
					dataSql += fmt.Sprintf(InsertMsgValSql, dataPayload.Id, dataPayload.Sender, dataPayload.Owner, dataPayload.ChatId, dataPayload.Content, time.Now().UnixMilli())
				}
			}

			execSql := fmt.Sprintf(InsertMsgSql, tableIndex, dataSql)
			_, err := m.Db.Execute(execSql)
			if err == nil {
				return nil
			}

			if mysqlErr, ok := err.(*mysql.MySQLError); ok {
				if mysqlErr.Number == 1146 {
					_, err := m.Db.Execute(fmt.Sprintf(CreateTableSql, dbIndex))
					if err == nil {
						_, err := m.Db.Execute(execSql)
						if err != nil {
							fn_log.Printf("%v", err)
						}
					}
				} else if mysqlErr.Number == 1150 {
					_, err := m.Db.Execute(execSql)
					if err != nil {
						fn_log.Printf("%v", err)
					}
				} else {
					fn_log.Printf("%v", err)
				}
			}
		}
	}

	return nil
}

func (m messageWriteDiffusionRepository) ListMessage(chatId int64, userId int64, lastTimestamp int64) (models []*MessageModel, err error) {
	return
}

type SaveMessageDbPayload struct {
	// int64-> tableIndex
	Payloads map[int64][]*SaveMessagePayload
}

type SaveMessagePayload struct {
	Id      int64
	ChatId  int64
	Sender  int64
	Owner   int64
	Content []byte
}

type messageWriteDiffusionRepository struct {
	Db sqlingo.Database `autowire:""` //过后优化
}

func NewMessageWriteDiffusionRepository() MessageWriteDiffusionRepository {
	return &messageWriteDiffusionRepository{}
}
