-- name: FindObservations :many
select
  *
from
  seen_observation
where
  project_id = sqlc.arg ('project_id')
  and channel_id = sqlc.arg ('channel_id')
  and id in (sqlc.slice ('id'));

-- name: CreateSeenObservation :one
insert
  or ignore into seen_observation (id, channel_id, project_id)
    values (?, ?, ?)
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
      channel_id = ?
      and message_id = ?
      and guild_id = ?
    limit 1);

-- name: SaveFeaturedMessage :one
insert
  or ignore into featured_message (message_id, channel_id, guild_id)
    values (?, ?, ?)
  returning
    *;

-- name: FindModuleConfiguration :one
select
  *
from
  module_configuration
where
  module = ?
  and key = ?;

-- name: FindModuleConfigurations :many
select
  *
from
  module_configuration
where
  module = ?;

-- name: CreateModuleConfiguration :one
insert into module_configuration (module, key, options)
  values (?, ?, ?)
returning
  *;

-- name: UpdateModuleConfiguration :one
update
  module_configuration
set
  options = ?
where
  key = ?
  and module = ?
returning
  *;

-- name: DeleteModuleConfiguration :one
delete from module_configuration
where module = ?
  and key = ?
returning
  *;
