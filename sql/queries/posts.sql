-- name: CreatePost :exec
insert into posts (
  id, feed_id, title, url, description, published_at, created_at, updated_at
) values 
($1, $2, $3, $4, $5, $6, $7, $8);