-- +goose Up
-- +goose StatementBegin
create table featured_message (
  channel_id text not null,
  message_id text not null,
  created_at timestamp default current_timestamp not null,
  updated_at timestamp default current_timestamp not null,
  primary key (channel_id, message_id)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table featured_message;

-- +goose StatementEnd
