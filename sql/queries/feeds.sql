-- name: CreateFeed :one
insert into feeds (id, user_id, name, url, created_at, updated_at) 
values (
$1,
$2, 
$3, 
$4,
$5, 
$6 ) returning *;

  