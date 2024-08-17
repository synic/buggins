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
