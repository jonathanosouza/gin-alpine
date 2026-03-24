-- name: CreateLink :one
insert into links (data, expires_at, uuid)
values ($1, $2, $3)
returning id;
-- name: CreateUserAvailableLinks :one
insert into user_available_links (user_id, link_uuid, type)
values ($1, $2, $3)
returning id;
-- name: GetLink :one
SELECT l.id,
    l.uuid,
    l.data,
    l.expires_at,
    l.created_at,
    l.updated_at
FROM links l
WHERE 1 = 1
    AND l.uuid = $1
    AND l.deleted_at is null
limit 1;
-- name: CheckIfUserHasResetPasswordLinksAvailable :one
select ual.link_uuid
from users u
    left join user_available_links ual on ual.user_id = u.id
    left join links l on l.uuid = ual.link_uuid
where 1 = 1
    and ual."type" = 'RESET_PASS'
    and l.deleted_at is null
    and l.expires_at > $2
    and u.email = $1
limit 1;
-- name: DeleteLink :exec
update links
set deleted_at = $2
where id = $1;
