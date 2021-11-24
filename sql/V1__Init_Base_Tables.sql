CREATE TABLE message (
     id BIGINT NOT NULL ,
     user_id BIGINT NOT NULL COMMENT 'sender',
     chat_id BIGINT NOT NULL COMMENT 'conversation id',
     content BLOB NOT NULL COMMENT 'message content' ,
     create_at TIMESTAMP DEFAULT current_timestamp(),
     PRIMARY KEY (`id`),
     INDEX m_user_id (`user_id`),
     INDEX m_chat_id_user_id (`chat_id` , `user_id`)
)

-- 分库 分表        1000w 人一个库        10万人一张表    100张表   库 userId % 1000w 选择库   userId % 10w 选择表
-- CREATE TABLE message_?(
--   id BIGINT NOT NULL ,
--   user_id BIGINT NOT NULL COMMENT 'sender',
--   chat_id BIGINT NOT NULL COMMENT 'conversation id',
--   content BLOB NOT NULL COMMENT 'message content' ,
--   create_at TIMESTAMP DEFAULT current_timestamp(),
--   PRIMARY KEY (`id`),
--   INDEX m_user_id (`user_id`),
--   INDEX m_chat_id_user_id (`chat_id` , `user_id`)
-- )