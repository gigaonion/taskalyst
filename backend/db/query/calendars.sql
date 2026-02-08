-- name: CreateCalendar :one
INSERT INTO calendars (
    user_id, name, color, description, project_id
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: ListCalendars :many
SELECT * FROM calendars
WHERE user_id = $1
ORDER BY created_at;

-- name: GetDefaultCalendar :one
SELECT * FROM calendars
WHERE user_id = $1
ORDER BY created_at
LIMIT 1;

-- name: CreateEvent :one
INSERT INTO scheduled_events (
    user_id, project_id, calendar_id,
    title, description, location,
    start_at, end_at, is_all_day,
    ical_uid, status, rrule, etag, sequence
) VALUES (
    $1, $2, $3,
    $4, $5, $6,
    $7, $8, $9,
    $10, $11, $12, $13, $14
) RETURNING *;

-- name: ListEventsByRange :many
SELECT e.*, p.title as project_title, p.category_id
FROM scheduled_events e
JOIN projects p ON e.project_id = p.id
WHERE
    e.user_id = $1
    AND e.end_at >= sqlc.arg('start_time')
    AND e.start_at <= sqlc.arg('end_time')
ORDER BY e.start_at ASC;

-- name: GetEventByICalUID :one
SELECT * FROM scheduled_events
WHERE user_id = $1 AND ical_uid = $2 LIMIT 1;
-- name: CreateTimetableSlot :one
INSERT INTO timetable_slots (
    user_id, project_id, day_of_week, start_time, end_time, location, note
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: ListTimetableSlots :many
SELECT ts.*, p.title as project_title, COALESCE(p.color, '#808080')::varchar as project_color
FROM timetable_slots ts
JOIN projects p ON ts.project_id = p.id
WHERE ts.user_id = $1
ORDER BY ts.day_of_week, ts.start_time;

-- name: ListTimetableSlotsByDayOfWeek :many
SELECT ts.*, p.title as project_title, COALESCE(p.color, '#808080')::varchar as project_color
FROM timetable_slots ts
JOIN projects p ON ts.project_id = p.id
WHERE ts.user_id = $1 AND ts.day_of_week = $2
ORDER BY ts.start_time;

-- name: ListEventsByCalendar :many
SELECT * FROM scheduled_events
WHERE user_id = $1 AND calendar_id = $2
ORDER BY start_at ASC;

-- name: ListEventsByCalendarAndRange :many
SELECT * FROM scheduled_events
WHERE user_id = $1 
  AND calendar_id = $2
  AND end_at >= sqlc.arg('start_time')
  AND start_at <= sqlc.arg('end_time')
ORDER BY start_at ASC;

-- name: GetCalendar :one
SELECT * FROM calendars
WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: UpdateEventByICalUID :one
UPDATE scheduled_events
SET
    title = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    location = COALESCE(sqlc.narg('location'), location),
    start_at = COALESCE(sqlc.narg('start_at'), start_at),
    end_at = COALESCE(sqlc.narg('end_at'), end_at),
    is_all_day = COALESCE(sqlc.narg('is_all_day'), is_all_day),
    status = COALESCE(sqlc.narg('status'), status),
    rrule = COALESCE(sqlc.narg('rrule'), rrule),
    etag = COALESCE(sqlc.narg('etag'), etag),
    sequence = COALESCE(sqlc.narg('sequence'), sequence),
    updated_at = NOW()
WHERE user_id = $1 AND ical_uid = $2
RETURNING *;

-- name: DeleteEventByICalUID :exec
DELETE FROM scheduled_events
WHERE user_id = $1 AND ical_uid = $2;

-- name: DeleteCalendar :exec
DELETE FROM calendars
WHERE id = $1 AND user_id = $2;
