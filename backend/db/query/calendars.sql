-- name: CreateCalendar :one
INSERT INTO calendars (
    user_id, name, color, description
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: ListCalendars :many
SELECT * FROM calendars
WHERE user_id = $1
ORDER BY created_at;

-- name: CreateEvent :one
INSERT INTO scheduled_events (
    user_id, project_id, calendar_id,
    title, description, location,
    start_at, end_at, is_all_day,
    ical_uid, status, rrule
) VALUES (
    $1, $2, $3,
    $4, $5, $6,
    $7, $8, $9,
    $10, $11, $12
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
SELECT ts.*, p.title as project_title, p.color as project_color
FROM timetable_slots ts
JOIN projects p ON ts.project_id = p.id
WHERE ts.user_id = $1
ORDER BY ts.day_of_week, ts.start_time;
