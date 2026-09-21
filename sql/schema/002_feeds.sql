-- +goose Up
create table feeds (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references users(id) on delete cascade,
  name text not null,
  url text unique not null,
  created_at timestamp not null,
  updated_at timestamp not null
);

-- +goose Down
drop table feeds;