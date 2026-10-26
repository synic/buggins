-- name: FindObservations :many
select
  *
from
  seen_observation
where
  project_id = sqlc.arg ('project_id')
  and channel_id = sqlc.arg ('channel_id')
  and id = any (sqlc.arg ('id')::bigint[]);

-- name: CreateSeenObservation :one
insert into seen_observation (id, channel_id, project_id)
  values ($1, $2, $3)
on conflict
  do nothing
returning
  *;

-- name: FindIsMessageFeatured :one
select
  exists (
    select
      1
    from
      featured_message
    where
      channel_id = $1
      and message_id = $2
      and guild_id = $3
    limit 1);

-- name: SaveFeaturedMessage :one
insert into featured_message (message_id, channel_id, guild_id)
  values ($1, $2, $3)
on conflict
  do nothing
returning
  *;

-- name: FindModuleConfiguration :one
select
  *
from
  module_configuration
where
  module = $1
  and key = $2;

-- name: FindModuleConfigurations :many
select
  *
from
  module_configuration
where
  module = $1;

-- name: CreateModuleConfiguration :one
insert into module_configuration (module, key, data)
  values ($1, $2, $3)
returning
  *;

-- name: UpdateModuleConfiguration :one
update
  module_configuration
set
  data = $1
where
  key = $2
  and module = $3
returning
  *;

-- name: DeleteModuleConfiguration :one
delete from module_configuration
where module = $1
  and key = $2
returning
  *;
