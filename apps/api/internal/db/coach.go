package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type CoachMessage struct {
	Role    string `json:"role"` // coach | user | system
	Content string `json:"content"`
}

type CoachSession struct {
	ID             string         `json:"id"`
	UserID         string         `json:"userId"`
	Scene          string         `json:"scene"` // job_search | promotion | communication
	Title          string         `json:"title"`
	RelatedTaskID  string         `json:"relatedTaskId,omitempty"`
	RelatedEvent   string         `json:"relatedEvent,omitempty"`
	Messages       []CoachMessage `json:"messages"`
	ActionItems    []string       `json:"actionItems"`
	Scripts        []string       `json:"scripts"`
	CrisisFlag     bool           `json:"crisisFlag"`
	Status         string         `json:"status"` // active | done
	CreatedAt      string         `json:"createdAt"`
	UpdatedAt      string         `json:"updatedAt"`
}

func (s *Store) migrateCoachWellbeing() error {
	const schema = `
CREATE TABLE IF NOT EXISTS coach_sessions (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  scene TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL DEFAULT '',
  related_task_id TEXT NOT NULL DEFAULT '',
  related_event TEXT NOT NULL DEFAULT '',
  messages TEXT,
  action_items TEXT,
  scripts TEXT,
  crisis_flag INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_coach_user_updated ON coach_sessions(user_id, updated_at);
CREATE TABLE IF NOT EXISTS wellbeing_checkins (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  at TEXT NOT NULL,
  stress_score INTEGER NOT NULL DEFAULT 0,
  mood_tags TEXT,
  energy_score INTEGER NOT NULL DEFAULT 0,
  event_type TEXT NOT NULL DEFAULT '',
  note TEXT NOT NULL DEFAULT '',
  related_task_id TEXT NOT NULL DEFAULT '',
  related_coach_session_id TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_wellbeing_user_at ON wellbeing_checkins(user_id, at);
`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) ListCoachSessions(userID string, limit int) ([]CoachSession, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
SELECT id, user_id, scene, title, related_task_id, related_event, messages, action_items, scripts,
       crisis_flag, status, created_at, updated_at
FROM coach_sessions WHERE user_id = ?
ORDER BY updated_at DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CoachSession, 0)
	for rows.Next() {
		sess, err := scanCoach(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sess)
	}
	return out, rows.Err()
}

func (s *Store) GetCoachSession(id, userID string) (*CoachSession, error) {
	row := s.db.QueryRow(`
SELECT id, user_id, scene, title, related_task_id, related_event, messages, action_items, scripts,
       crisis_flag, status, created_at, updated_at
FROM coach_sessions WHERE id = ? AND user_id = ?`, id, userID)
	sess, err := scanCoach(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *Store) CreateCoachSession(sess *CoachSession) error {
	_, err := s.db.Exec(`
INSERT INTO coach_sessions (
  id, user_id, scene, title, related_task_id, related_event, messages, action_items, scripts,
  crisis_flag, status, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sess.ID, sess.UserID, sess.Scene, sess.Title, sess.RelatedTaskID, sess.RelatedEvent,
		mustJSON(sess.Messages), mustJSON(sess.ActionItems), mustJSON(sess.Scripts),
		boolInt(sess.CrisisFlag), sess.Status, sess.CreatedAt, sess.UpdatedAt,
	)
	return err
}

func (s *Store) UpdateCoachSession(sess *CoachSession) error {
	res, err := s.db.Exec(`
UPDATE coach_sessions SET
  scene = ?, title = ?, related_task_id = ?, related_event = ?, messages = ?,
  action_items = ?, scripts = ?, crisis_flag = ?, status = ?, updated_at = ?
WHERE id = ? AND user_id = ?`,
		sess.Scene, sess.Title, sess.RelatedTaskID, sess.RelatedEvent, mustJSON(sess.Messages),
		mustJSON(sess.ActionItems), mustJSON(sess.Scripts), boolInt(sess.CrisisFlag), sess.Status, sess.UpdatedAt,
		sess.ID, sess.UserID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("session not found")
	}
	return nil
}

func (s *Store) DeleteCoachSession(id, userID string) error {
	res, err := s.db.Exec(`DELETE FROM coach_sessions WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("session not found")
	}
	return nil
}

func scanCoach(sc rowScanner) (CoachSession, error) {
	var sess CoachSession
	var messages, actions, scripts sql.NullString
	var crisis int
	err := sc.Scan(
		&sess.ID, &sess.UserID, &sess.Scene, &sess.Title, &sess.RelatedTaskID, &sess.RelatedEvent,
		&messages, &actions, &scripts, &crisis, &sess.Status, &sess.CreatedAt, &sess.UpdatedAt,
	)
	if err != nil {
		return CoachSession{}, err
	}
	sess.CrisisFlag = crisis != 0
	sess.Messages = decodeSlice[CoachMessage](messages)
	sess.ActionItems = decodeStringSlice(actions)
	sess.Scripts = decodeStringSlice(scripts)
	if sess.Messages == nil {
		sess.Messages = []CoachMessage{}
	}
	if sess.ActionItems == nil {
		sess.ActionItems = []string{}
	}
	if sess.Scripts == nil {
		sess.Scripts = []string{}
	}
	return sess, nil
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil || string(b) == "null" {
		return "[]"
	}
	return string(b)
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func decodeSlice[T any](ns sql.NullString) []T {
	if !ns.Valid || ns.String == "" || ns.String == "null" {
		return nil
	}
	var out []T
	if err := json.Unmarshal([]byte(ns.String), &out); err != nil {
		return nil
	}
	return out
}

func decodeStringSlice(ns sql.NullString) []string {
	return decodeSlice[string](ns)
}
