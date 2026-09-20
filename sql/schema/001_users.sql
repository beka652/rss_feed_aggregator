-- +goose Up
create table users(
  id uuid primary key default gen_random_uuid(),
  name text unique not null, 
  created_at timestamp not null, 
  updated_at timestamp not null
);
-- +goose Down 
drop table users;