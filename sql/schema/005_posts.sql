-- +goose Up
create table posts (
  id uuid primary key default gen_random_uuid(),
  feed_id uuid not null references feeds(id) on delete cascade,
  title text not null,
  url text unique not null, 
  description text not null,
  published_at timestamp,
  created_at timestamp not null,
  updated_at timestamp not null 
);

-- +goose Down 
drop table posts;