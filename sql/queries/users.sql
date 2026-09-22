-- name: CreateUser :one
insert into users (id, created_at, updated_at, name) 
values (
$1,
$2,
$3,
$4
)
returning *; 

-- name: GetUser :one
select * from users where name = $1 limit 1;
-- name: ResetDB :exec 
delete from users;
-- name: GetUsers :many
select * from users;
-- name: GetPostsForUser :many 
select 
  * 
from 
  posts 
where feed_id in (
  select 
    feed_id
  from 
    feed_follows 
  where user_id = $1
  )
order by published_at desc nulls last
limit $2;