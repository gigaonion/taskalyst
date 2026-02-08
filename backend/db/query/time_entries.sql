-- name: CreateTimeEntry :one
INSERT INTO time_entries (
  user_id, project_id, task_id, started_at,ended_at, note
) VALUES (
    $1, $2, $3, $4, $5 ,$6
) RETURNING *;

-- name: StopTimeEntry :one
UPDATE time_entries
SET ended_at = $3, updated_at = NOW()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: UpdateTimeEntry :one
UPDATE time_entries
SET 
    project_id = COALESCE(sqlc.narg('project_id'), project_id),
    task_id = COALESCE(sqlc.narg('task_id'), task_id),
    started_at = COALESCE(sqlc.narg('started_at'), started_at),
    ended_at = COALESCE(sqlc.narg('ended_at'), ended_at),
    note = COALESCE(sqlc.narg('note'), note)
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: GetRunningTimeEntries :many
-- 計測中のエントリ
SELECT t.*, p.title as project_title, p.color as project_color
FROM time_entries t
JOIN projects p ON t.project_id = p.id
WHERE t.user_id = $1 AND t.ended_at IS NULL
ORDER BY t.started_at DESC;

-- name: ListTimeEntries :many
SELECT t.*, p.title as project_title, p.color as project_color
FROM time_entries t
JOIN projects p ON t.project_id = p.id
WHERE 
    t.user_id = $1
    AND t.started_at >= @from_date
    AND t.started_at <= @to_date
ORDER BY t.started_at DESC;

-- name: GetGrowthStats :many
-- GROWTHカテゴリの実績のみを日別集計
SELECT
    DATE(te.started_at)::text as date,
    SUM(EXTRACT(EPOCH FROM (COALESCE(te.ended_at, NOW()) - te.started_at)))::bigint as total_seconds
FROM time_entries te
JOIN projects p ON te.project_id = p.id
JOIN categories c ON p.category_id = c.id
WHERE
    te.user_id = $1
    AND c.root_type = 'GROWTH'
    AND te.started_at >= @from_date
    AND te.started_at <= @to_date
GROUP BY DATE(te.started_at)
ORDER BY date;
