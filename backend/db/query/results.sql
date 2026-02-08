-- name: CreateResult :one
INSERT INTO results (
    user_id, project_id, target_task_id, type, value, recorded_at, note
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: ListResults :many
SELECT 
    r.*, 
    p.title as project_title, 
    p.color as project_color, 
    t.title as task_title
FROM results r
JOIN projects p ON r.project_id = p.id
LEFT JOIN tasks t ON r.target_task_id = t.id
WHERE
    r.user_id = $1
    AND (sqlc.narg('project_id')::uuid IS NULL OR r.project_id = @project_id)
    AND r.recorded_at >= @from_date
    AND r.recorded_at <= @to_date
ORDER BY r.recorded_at DESC;

-- name: DeleteResult :exec
DELETE FROM results
WHERE id = $1 AND user_id = $2;
