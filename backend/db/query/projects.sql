-- name: CreateProject :one
INSERT INTO projects (
    user_id, category_id, title, description, color, is_archived
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetProject :one
SELECT * FROM projects
WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListProjects :many
SELECT 
    p.id, p.user_id, p.category_id, p.title, p.description, 
    COALESCE(p.color, '#808080')::varchar as color, 
    p.is_archived, p.created_at, p.updated_at, 
    c.name as category_name, c.root_type, 
    COALESCE(c.color, '#808080')::varchar as category_color
FROM projects p
JOIN categories c ON p.category_id = c.id
WHERE
    p.user_id = $1
    AND (sqlc.narg('is_archived')::boolean IS NULL OR p.is_archived = @is_archived)
ORDER BY p.updated_at DESC;

-- name: GetDefaultProject :one
SELECT * FROM projects
WHERE user_id = $1
ORDER BY created_at ASC
LIMIT 1;

-- name: CreateCategory :one
INSERT INTO categories (
    user_id, name, root_type, color
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: ListCategories :many
SELECT * FROM categories
WHERE user_id = $1
ORDER BY root_type, name;
