package db

const schema = `
CREATE TABLE IF NOT EXISTS boards (
	id          TEXT PRIMARY KEY,
	name        TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	created_at  DATETIME NOT NULL,
	updated_at  DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS columns (
	id         TEXT PRIMARY KEY,
	board_id   TEXT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
	name       TEXT NOT NULL,
	position   INTEGER NOT NULL DEFAULT 0,
	wip_limit  INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks (
	id           TEXT PRIMARY KEY,
	board_id     TEXT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
	column_id    TEXT NOT NULL REFERENCES columns(id) ON DELETE CASCADE,
	title        TEXT NOT NULL,
	description  TEXT NOT NULL DEFAULT '',
	priority     TEXT NOT NULL DEFAULT 'medium',
	labels       TEXT NOT NULL DEFAULT '',
	position     INTEGER NOT NULL DEFAULT 0,
	created_at   DATETIME NOT NULL,
	updated_at   DATETIME NOT NULL,
	completed_at DATETIME,
	due_date     DATETIME
);

CREATE TABLE IF NOT EXISTS comments (
	id         TEXT PRIMARY KEY,
	task_id    TEXT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
	author     TEXT NOT NULL DEFAULT 'claude',
	body       TEXT NOT NULL,
	created_at DATETIME NOT NULL
);

PRAGMA foreign_keys = ON;
`
