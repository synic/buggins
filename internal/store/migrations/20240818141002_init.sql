-- +goose Up
create table seen_observation (
  id integer primary key,
  created_at timestamp default current_timestamp not null,
  updated_at timestamp default current_timestamp not null
);

-- +goose Down
drop table seen_observation;
