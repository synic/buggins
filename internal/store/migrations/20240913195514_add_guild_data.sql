-- +goose Up
-- +goose StatementBegin
create table seen_obs_copy (
  id integer not null,
  channel_id text not null,
  project_id integer not null,
  created_at timestamp default current_timestamp not null,
  updated_at timestamp default current_timestamp not null,
  primary key (id, channel_id, project_id)
);

insert into seen_obs_copy (id, channel_id, project_id, created_at, updated_at)
select
  id,
  '1091172345760186378',
  146454,
  created_at,
  updated_at
from
  seen_observation;

drop table seen_observation;

alter table seen_obs_copy rename to seen_observation;

create table feat_msg_copy (
  guild_id text not null,
  channel_id text not null,
  message_id text not null,
  created_at timestamp default current_timestamp not null,
  updated_at timestamp default current_timestamp not null,
  primary key (guild_id, channel_id, message_id)
);

insert into feat_msg_copy (guild_id, channel_id, message_id, created_at, updated_at)
select
  '1016730463710232627',
  channel_id,
  message_id,
  created_at,
  updated_at
from
  featured_message;

drop table featured_message;

alter table feat_msg_copy rename to featured_message;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
create table seen_obs_copy (
  id integer primary key,
  created_at timestamp default current_timestamp not null,
  updated_at timestamp default current_timestamp not null
);

insert into seen_obs_copy (id, created_at, updated_at)
select
  id,
  created_at,
  updated_at
from
  seen_observation;

drop table seen_observation;

alter table seen_obs_copy rename to seen_observation;

create table feat_msg_copy (
  channel_id text not null,
  message_id text not null,
  created_at timestamp default current_timestamp not null,
  updated_at timestamp default current_timestamp not null,
  primary key (channel_id, message_id)
);

insert into feat_msg_copy (channel_id, message_id, created_at, updated_at)
select
  channel_id,
  message_id,
  created_at,
  updated_at
from
  featured_message;

drop table featured_message;

alter table feat_msg_copy rename to featured_message;

-- +goose StatementEnd
