-- Enable UUID --
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
-- Immute Data --
CREATE TYPE user_role AS ENUM('ADMIN', 'USER');
CREATE TYPE task_status AS ENUM('TODO','DOING','DONE');
CREATE TYPE root_category_type AS ENUM('GROWTH','LIFE','WORK','HOBBY','OTHER');
-- Table --
-- user
CREATE TABLE users(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  name VARCHAR(100) NOT NULL,
  role user_role NOT NULL DEFAULT 'USER',
  preferences JSONB DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
--API
CREATE TABLE api_tokens(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name VARCHAR(100) NOT NULL,
  token_hash VARCHAR(64) NOT NULL,
  expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
--category
CREATE TABLE categories(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name VARCHAR(50) NOT NULL,
  root_type root_category_type NOT NULL,
  color VARCHAR(7) DEFAULT '#808080',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id,name)
);
-- project
CREATE TABLE projects(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  category_id UUID NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
  title VARCHAR(100) NOT NULL,
  description TEXT,
  is_archived BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
--task
CREATE TABLE tasks(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  title VARCHAR(200) NOT NULL,
  note_markdown TEXT,
  status task_status NOT NULL DEFAULT 'TODO',
  due_date TIMESTAMPTZ,
  priority SMALLINT DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  -- for caldav
  calendar_id UUID REFERENCES calendars(id) ON DELETE CASCADE,
  ical_uid VARCHAR(255),
  etag VARCHAR(64),
  sequence INTEGER NOT NULL DEFAULT 0,
  completed_at TIMESTAMPTZ
);
-- child task
CREATE TABLE checklist_items(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id UUID NOT NULL REFERENCES task(id) ON DELETE CASCADE,
  content VARCHAR(255) NOT NULL,
  is_completed BOOLEAN NOT NULL DEFAULT FALSE,
  position INTEGER NOT NULL DEFAULT 0
);
-- time table
CREATE TABLE timetable_slots(
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  day_of_week SMALLINT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
  start_time TIME NOT NULL,
  end_time TIME NOT NULL,
  note TEXT,
  location VARCHAR(100),
  EXCLUDE USING gist (
      user_id WITH =,
      day_of_week WITH =,
      tsrange(
          ('2000-01-01'::date + start_time)::timestamp,
          ('2000-01-01'::date + end_time)::timestamp
      ) WITH &&
  )
);
-- calendar
CREATE TABLE calendars (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7),
    description TEXT,
    sync_token VARCHAR(255) NOT NULL DEFAULT '1',
    supported_components VARCHAR(50)[] DEFAULT ARRAY['VEVENT', 'VTODO'],
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE scheduled_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    calendar_id UUID REFERENCES calendars(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    location VARCHAR(100),
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    is_all_day BOOLEAN NOT NULL DEFAULT FALSE,
    external_event_id VARCHAR(255),
    -- for ical
    ical_uid VARCHAR(255),
    etag VARCHAR(64),
    sequence INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20),
    transparency VARCHAR(20) DEFAULT 'OPAQUE',
    rrule TEXT,
    dtstamp TIMESTAMPTZ DEFAULT NOW(),
    url TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_event_duration CHECK (end_at > start_at)
);
CREATE TABLE time_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    note TEXT,
    is_auto_generated BOOLEAN NOT NULL DEFAULT FALSE,
    
    CONSTRAINT valid_duration CHECK (ended_at IS NULL OR ended_at > started_at)
);
-- achievement
CREATE TABLE results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    target_task_id UUID REFERENCES tasks(id) ON DELETE SET NULL,
    
    type VARCHAR(50) NOT NULL,
    value NUMERIC(12, 2) NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    note TEXT
);
-- performance
-- Categorie
CREATE INDEX idx_categories_root_type ON categories(user_id, root_type);
-- project
CREATE INDEX idx_projects_active ON projects(user_id) WHERE is_archived = FALSE;
-- task
CREATE INDEX idx_tasks_active_user ON tasks(user_id, due_date) WHERE status != 'DONE';
CREATE INDEX idx_tasks_project ON tasks(project_id);
CREATE INDEX idx_tasks_ical_uid ON tasks(ical_uid);
-- calendar
CREATE INDEX idx_scheduled_events_range ON scheduled_events (user_id, start_at, end_at);
CREATE INDEX idx_scheduled_events_ical_uid ON scheduled_events(ical_uid);
-- time
CREATE INDEX idx_time_entries_range ON time_entries(user_id, started_at DESC);
CREATE INDEX idx_time_entries_project ON time_entries(project_id, started_at DESC);
CREATE INDEX idx_results_user_date ON results(user_id, recorded_at DESC);
