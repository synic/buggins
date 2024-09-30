-- +goose Up
-- +goose StatementBegin
create table module_configuration (
  module text not null,
  key text not null,
  data json not null default '{}',
  primary key (module, key)
);

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
drop table module_configuration;

-- +goose StatementEnd
