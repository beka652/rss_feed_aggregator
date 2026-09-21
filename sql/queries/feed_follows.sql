-- name: CreateFeedFollow :one
with inserted_feed_follow as (
  insert into feed_follows (id, user_id, feed_id, created_at, updated_at) 
  values ($1, $2, $3, $4,$5) returning *
)
select 
  feeds.name as feed_name,
  users.name as user_name,
  inserted_feed_follow.*
from inserted_feed_follow
inner join feeds on inserted_feed_follow.feed_id = feeds.id
inner join users on inserted_feed_follow.user_id = users.id
;

-- name: GetFeedFollowsForUser :many
select 
  feeds.name AS feed_name, 
  feed_follows.*
from feed_follows
inner join feeds on feed_follows.feed_id = feeds.id
where feed_follows.user_id = $1;

-- name: DeleteFeedFollowByUrl :exec
delete from feed_follows 
where feed_follows.user_id = $1 and 
  feed_id = (select id from feeds where url = $2);



  