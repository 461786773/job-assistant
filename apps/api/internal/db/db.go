package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db   *sql.DB
	path string
}

type Task struct {
	ID             string          `json:"id"`
	UserID         string          `json:"userId,omitempty"`
	Title          string          `json:"title"`
	Company        string          `json:"company"`
	TargetRole     string          `json:"targetRole"`
	JDText         string          `json:"jdText"`
	ResumeText     string          `json:"resumeText"`
	ResumeFilename string          `json:"resumeFilename"`
	ResumeFormat   string          `json:"resumeFormat"`
	Status         string          `json:"status"`
	Notes          string          `json:"notes"`
	HrReport       json.RawMessage `json:"hrReport,omitempty"`
	Interview      json.RawMessage `json:"interview,omitempty"`
	Salary         json.RawMessage `json:"salary,omitempty"`
	CreatedAt      string          `json:"createdAt"`
	UpdatedAt      string          `json:"updatedAt"`
}

type diskFormat struct {
	Tasks []Task `json:"tasks"`
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	dsn := path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)"
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	s := &Store{db: sqlDB, path: path}
	if err := s.Migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	if err := s.importJSONIfNeeded(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("import tasks.json: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Migrate() error {
	const schema = `
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS tasks (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL DEFAULT '',
  company TEXT NOT NULL DEFAULT '',
  target_role TEXT NOT NULL DEFAULT '',
  jd_text TEXT NOT NULL DEFAULT '',
  resume_text TEXT NOT NULL DEFAULT '',
  resume_filename TEXT NOT NULL DEFAULT '',
  resume_format TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT '',
  notes TEXT NOT NULL DEFAULT '',
  hr_report TEXT,
  interview TEXT,
  salary TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tasks_updated_at ON tasks(updated_at);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
`
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := s.ensureColumn("tasks", "user_id", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	if _, err := s.db.Exec(`CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id)`); err != nil {
		return fmt.Errorf("migrate index: %w", err)
	}
	if err := s.migrateAssessmentBooking(); err != nil {
		return fmt.Errorf("migrate assessment/booking: %w", err)
	}
	if err := s.migrateBigFive(); err != nil {
		return fmt.Errorf("migrate bigfive: %w", err)
	}
	if err := s.seedDefaultUser(); err != nil {
		return fmt.Errorf("seed default user: %w", err)
	}
	if err := s.migrateCoachWellbeing(); err != nil {
		return fmt.Errorf("migrate coach/wellbeing: %w", err)
	}
	return nil
}

func (s *Store) ensureColumn(table, column, decl string) error {
	rows, err := s.db.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = s.db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, table, column, decl))
	return err
}

func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

const taskSelectCols = `id, user_id, title, company, target_role, jd_text, resume_text, resume_filename, resume_format,
       status, notes, hr_report, interview, salary, created_at, updated_at`

func (s *Store) ListTasks(userID string) ([]Task, error) {
	rows, err := s.db.Query(`
SELECT `+taskSelectCols+`
FROM tasks
WHERE user_id = ?
ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Task, 0)
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) GetTask(id, userID string) (*Task, error) {
	row := s.db.QueryRow(`
SELECT `+taskSelectCols+`
FROM tasks WHERE id = ? AND user_id = ?`, id, userID)
	t, err := scanTask(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) CreateTask(t *Task) error {
	_, err := s.db.Exec(`
INSERT INTO tasks (
  id, user_id, title, company, target_role, jd_text, resume_text, resume_filename, resume_format,
  status, notes, hr_report, interview, salary, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.UserID, t.Title, t.Company, t.TargetRole, t.JDText, t.ResumeText, t.ResumeFilename, t.ResumeFormat,
		t.Status, t.Notes, nullJSON(t.HrReport), nullJSON(t.Interview), nullJSON(t.Salary),
		t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func (s *Store) UpdateTask(t *Task) error {
	res, err := s.db.Exec(`
UPDATE tasks SET
  title = ?, company = ?, target_role = ?, jd_text = ?, resume_text = ?,
  resume_filename = ?, resume_format = ?, status = ?, notes = ?,
  hr_report = ?, interview = ?, salary = ?, updated_at = ?
WHERE id = ? AND user_id = ?`,
		t.Title, t.Company, t.TargetRole, t.JDText, t.ResumeText,
		t.ResumeFilename, t.ResumeFormat, t.Status, t.Notes,
		nullJSON(t.HrReport), nullJSON(t.Interview), nullJSON(t.Salary), t.UpdatedAt,
		t.ID, t.UserID,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

func (s *Store) DeleteTask(id, userID string) error {
	res, err := s.db.Exec(`DELETE FROM tasks WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(sc rowScanner) (Task, error) {
	var t Task
	var hr, interview, salary sql.NullString
	err := sc.Scan(
		&t.ID, &t.UserID, &t.Title, &t.Company, &t.TargetRole, &t.JDText, &t.ResumeText,
		&t.ResumeFilename, &t.ResumeFormat, &t.Status, &t.Notes,
		&hr, &interview, &salary, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return Task{}, err
	}
	t.HrReport = rawJSON(hr)
	t.Interview = rawJSON(interview)
	t.Salary = rawJSON(salary)
	return t, nil
}

func nullJSON(v json.RawMessage) any {
	if len(v) == 0 || string(v) == "null" {
		return nil
	}
	return string(v)
}

func rawJSON(ns sql.NullString) json.RawMessage {
	if !ns.Valid || strings.TrimSpace(ns.String) == "" {
		return nil
	}
	return json.RawMessage(ns.String)
}

// importJSONIfNeeded migrates legacy data/tasks.json into SQLite once.
func (s *Store) importJSONIfNeeded() error {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	jsonPath := filepath.Join(filepath.Dir(s.path), "tasks.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}
	var disk diskFormat
	if err := json.Unmarshal(data, &disk); err != nil {
		return err
	}
	if len(disk.Tasks) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(`
INSERT OR IGNORE INTO tasks (
  id, user_id, title, company, target_role, jd_text, resume_text, resume_filename, resume_format,
  status, notes, hr_report, interview, salary, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, t := range disk.Tasks {
		if _, err := stmt.Exec(
			t.ID, t.UserID, t.Title, t.Company, t.TargetRole, t.JDText, t.ResumeText, t.ResumeFilename, t.ResumeFormat,
			t.Status, t.Notes, nullJSON(t.HrReport), nullJSON(t.Interview), nullJSON(t.Salary),
			t.CreatedAt, t.UpdatedAt,
		); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	bak := jsonPath + ".migrated"
	_ = os.Rename(jsonPath, bak)
	return nil
}
