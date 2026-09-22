-- name: CreateFeed :one
insert into feeds (id, user_id, name, url, created_at, updated_at) 
values (
$1,
$2, 
$3, 
$4,
$5, 
$6 ) returning *;

-- name: GetFeeds :many
select 
  feeds.name as feed_name,
  feeds.url as url,
  users.name as user_name
from feeds 
inner join users on feeds.user_id = users.id;
-- name: GetFeedByUrl :one 
select * from feeds where url=$1; 

-- name: MarkFeedFetched :exec
update  feeds 
set last_fetched_at = $2, updated_at = $3
where feeds.id = $1
;

-- name: GetNextFeedToFetch :one 
select 
  id,
  name,
  url
from feeds 
order by last_fetched_at asc nulls first
limit 1;