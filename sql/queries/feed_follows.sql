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
WITH queried_user AS (
  SELECT *  FROM users WHERE users.name = $1
)
SELECT 
  qu.name AS user_name, 
  feeds.name AS feed_name, 
  feed_follows.*
FROM feed_follows 
INNER JOIN queried_user qu ON feed_follows.user_id = qu.id 
INNER JOIN feeds ON feed_follows.feed_id = feeds.id;

  