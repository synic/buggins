-- name: FindObservationsByIds :many
select
  *
from
  seen_observation
where
  id in (sqlc.slice ('ids'));

-- name: CreateSeenObservation :one
insert
  or ignore into seen_observation (id)
    values (?)
  returning
    *;

-- name: IsMessageFeatured :one
select
  exists (
    select
      1
    from
      featured_message
    where
      channel_id = ?
      and message_id = ?
    limit 1);

-- name: SaveFeaturedMessage :one
insert
  or ignore into featured_message (message_id, channel_id)
    values (?, ?)
  returning
    *;
