-- Migration: Create tasks table
-- Version: 2
-- Description: Add task list management for estimation tracking

CREATE TABLE IF NOT EXISTS tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id VARCHAR(10) NOT NULL,
    headline VARCHAR(255) NOT NULL,
    description TEXT,
    tracker_link TEXT,
    estimation VARCHAR(10),
    position INTEGER NOT NULL,
    CONSTRAINT fk_room FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    CONSTRAINT unique_room_position UNIQUE (room_id, position)
);

CREATE INDEX IF NOT EXISTS idx_tasks_room_id ON tasks(room_id);
CREATE INDEX IF NOT EXISTS idx_tasks_position ON tasks(room_id, position);
