-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    telegram_message_id BIGINT NOT NULL,
    telegram_chat_id BIGINT NOT NULL,
    telegram_user_id BIGINT NOT NULL,
    text TEXT,
    direction VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    is_command BOOLEAN DEFAULT FALSE,
    command_name VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    
    UNIQUE(telegram_chat_id, telegram_message_id)
);

CREATE INDEX idx_messages_chat_id ON messages(telegram_chat_id);
CREATE INDEX idx_messages_user_id ON messages(telegram_user_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);
CREATE INDEX idx_messages_command ON messages(is_command, command_name) WHERE is_command = true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS messages;
-- +goose StatementEnd
