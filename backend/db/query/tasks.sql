-- name: CreateTask :one
INSERT INTO tasks (
    user_id, project_id, title, note_markdown, due_date, priority,
    calendar_id, ical_uid, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetTask :one
SELECT * FROM tasks
WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListTasksWithStats :many
-- タスクと同時に、チェックリストの進捗を取得
SELECT
    t.id, t.project_id, t.title, t.status, t.due_date, t.priority,
    p.title as project_title, p.color as project_color,
    COUNT(ci.id) as total_items,
    COUNT(ci.id) FILTER (WHERE ci.is_completed) as done_items
FROM tasks t
JOIN projects p ON t.project_id = p.id
LEFT JOIN checklist_items ci ON t.id = ci.task_id
WHERE
    t.user_id = $1
    AND (sqlc.narg('project_id')::uuid IS NULL OR t.project_id = @project_id)
    AND (sqlc.narg('status')::task_status IS NULL OR t.status = @status)
    AND (sqlc.narg('from_date')::timestamptz IS NULL OR t.due_date >= @from_date)
    AND (sqlc.narg('to_date')::timestamptz IS NULL OR t.due_date <= @to_date)
GROUP BY t.id, p.id
ORDER BY
    CASE WHEN t.status = 'DONE' THEN 1 ELSE 0 END,
    t.due_date ASC NULLS LAST,
    t.created_at DESC;

-- name: UpdateTask :one
UPDATE tasks
SET
    title = COALESCE(sqlc.narg('title'), title),
    note_markdown = COALESCE(sqlc.narg('note_markdown'), note_markdown),
    status = $3,
    completed_at = $4,
    due_date = COALESCE(sqlc.narg('due_date'), due_date),
    priority = COALESCE(sqlc.narg('priority'), priority),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: ListTasksByCalendar :many
SELECT * FROM tasks
WHERE user_id = $1 AND calendar_id = $2
ORDER BY created_at DESC;

-- name: ListTasksByCalendarAndRange :many
SELECT * FROM tasks
WHERE user_id = $1
  AND calendar_id = $2
  AND (due_date IS NULL OR (due_date >= sqlc.arg('start_time') AND due_date <= sqlc.arg('end_time')))
ORDER BY created_at DESC;

-- name: GetTaskByICalUID :one
SELECT * FROM tasks
WHERE user_id = $1 AND ical_uid = $2 LIMIT 1;

-- name: UpdateTaskByICalUID :one
UPDATE tasks
SET
    title = COALESCE(sqlc.narg('title'), title),
    note_markdown = COALESCE(sqlc.narg('note_markdown'), note_markdown),
    status = COALESCE(sqlc.narg('status'), status),
    due_date = COALESCE(sqlc.narg('due_date'), due_date),
    priority = COALESCE(sqlc.narg('priority'), priority),
    etag = COALESCE(sqlc.narg('etag'), etag),
    sequence = COALESCE(sqlc.narg('sequence'), sequence),
    completed_at = COALESCE(sqlc.narg('completed_at'), completed_at),
    updated_at = NOW()
WHERE user_id = $1 AND ical_uid = $2
RETURNING *;

-- name: DeleteTaskByICalUID :exec
DELETE FROM tasks
WHERE user_id = $1 AND ical_uid = $2;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1 AND user_id = $2;


-- name: CreateChecklistItem :one
INSERT INTO checklist_items (
    task_id, content, position
) VALUES (
    $1, $2,
    (SELECT COALESCE(MAX(position), 0) + 1 FROM checklist_items WHERE task_id = $1)
) RETURNING *;

-- name: ListChecklistItems :many
SELECT * FROM checklist_items
WHERE task_id = $1
ORDER BY position ASC;

-- name: UpdateChecklistItem :one
UPDATE checklist_items
SET 
    content = COALESCE(sqlc.narg('content'), content),
    is_completed = COALESCE(sqlc.narg('is_completed'), is_completed),
    position = COALESCE(sqlc.narg('position'), position)
WHERE id = $1
RETURNING *;

-- name: DeleteChecklistItem :exec
DELETE FROM checklist_items
WHERE id = $1;
