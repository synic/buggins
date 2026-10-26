-- +goose Up
-- +goose StatementBegin
create table seen_observation (
  id bigint not null,
  channel_id text not null,
  project_id bigint not null,
  created_at timestamptz default now() not null,
  updated_at timestamptz default now() not null,
  primary key (id, channel_id, project_id)
);

create table featured_message (
  guild_id text not null,
  channel_id text not null,
  message_id text not null,
  created_at timestamptz default now() not null,
  updated_at timestamptz default now() not null,
  primary key (guild_id, channel_id, message_id)
);

create table module_configuration (
  module text not null,
  key text not null,
  data jsonb not null default '{}',
  primary key (module, key)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table module_configuration;

drop table featured_message;

drop table seen_observation;

-- +goose StatementEnd
