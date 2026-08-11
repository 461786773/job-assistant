package db

import (
	"database/sql"
	"encoding/json"
)

type WorkplaceBigFiveProfile struct {
	ID              string          `json:"id"`
	UserID          string          `json:"userId"`
	Version         string          `json:"version"`
	RawAnswers      json.RawMessage `json:"rawAnswers"`
	Scores          json.RawMessage `json:"scores"`
	PersonaID       string          `json:"personaId"`
	PersonaTitle    string          `json:"personaTitle"`
	PersonaBlurb    string          `json:"personaBlurb"`
	PersonaBody     string          `json:"personaBody,omitempty"`
	Tags            []string        `json:"tags"`
	CoachHints      []string        `json:"coachHints"`
	SummaryForCoach string          `json:"summaryForCoach,omitempty"`
	CompletedAt     string          `json:"completedAt"`
	CreatedAt       string          `json:"createdAt"`
}

func (s *Store) migrateBigFive() error {
	const schema = `
CREATE TABLE IF NOT EXISTS workplace_bigfive_profiles (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  version TEXT NOT NULL DEFAULT 'v1',
  raw_answers TEXT,
  scores TEXT,
  persona_id TEXT NOT NULL DEFAULT '',
  persona_title TEXT NOT NULL DEFAULT '',
  persona_blurb TEXT NOT NULL DEFAULT '',
  persona_body TEXT NOT NULL DEFAULT '',
  tags TEXT,
  coach_hints TEXT,
  summary_for_coach TEXT NOT NULL DEFAULT '',
  completed_at TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_bigfive_user_completed ON workplace_bigfive_profiles(user_id, completed_at);
`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) CreateBigFiveProfile(p *WorkplaceBigFiveProfile) error {
	_, err := s.db.Exec(`
INSERT INTO workplace_bigfive_profiles (
  id, user_id, version, raw_answers, scores, persona_id, persona_title, persona_blurb, persona_body,
  tags, coach_hints, summary_for_coach, completed_at, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.UserID, p.Version, nullJSON(p.RawAnswers), nullJSON(p.Scores),
		p.PersonaID, p.PersonaTitle, p.PersonaBlurb, p.PersonaBody,
		mustJSON(p.Tags), mustJSON(p.CoachHints), p.SummaryForCoach, p.CompletedAt, p.CreatedAt,
	)
	return err
}

func (s *Store) LatestBigFiveProfile(userID string) (*WorkplaceBigFiveProfile, error) {
	row := s.db.QueryRow(`
SELECT id, user_id, version, raw_answers, scores, persona_id, persona_title, persona_blurb, persona_body,
       tags, coach_hints, summary_for_coach, completed_at, created_at
FROM workplace_bigfive_profiles WHERE user_id = ?
ORDER BY completed_at DESC LIMIT 1`, userID)
	p, err := scanBigFive(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) GetBigFiveProfile(id, userID string) (*WorkplaceBigFiveProfile, error) {
	row := s.db.QueryRow(`
SELECT id, user_id, version, raw_answers, scores, persona_id, persona_title, persona_blurb, persona_body,
       tags, coach_hints, summary_for_coach, completed_at, created_at
FROM workplace_bigfive_profiles WHERE id = ? AND user_id = ?`, id, userID)
	p, err := scanBigFive(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) ListBigFiveProfiles(userID string, limit int) ([]WorkplaceBigFiveProfile, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.Query(`
SELECT id, user_id, version, raw_answers, scores, persona_id, persona_title, persona_blurb, persona_body,
       tags, coach_hints, summary_for_coach, completed_at, created_at
FROM workplace_bigfive_profiles WHERE user_id = ?
ORDER BY completed_at DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]WorkplaceBigFiveProfile, 0)
	for rows.Next() {
		p, err := scanBigFive(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func scanBigFive(sc rowScanner) (WorkplaceBigFiveProfile, error) {
	var p WorkplaceBigFiveProfile
	var answers, scores, tags, hints sql.NullString
	err := sc.Scan(
		&p.ID, &p.UserID, &p.Version, &answers, &scores,
		&p.PersonaID, &p.PersonaTitle, &p.PersonaBlurb, &p.PersonaBody,
		&tags, &hints, &p.SummaryForCoach, &p.CompletedAt, &p.CreatedAt,
	)
	if err != nil {
		return p, err
	}
	if answers.Valid {
		p.RawAnswers = json.RawMessage(answers.String)
	}
	if scores.Valid {
		p.Scores = json.RawMessage(scores.String)
	}
	_ = json.Unmarshal([]byte(nullStr(tags)), &p.Tags)
	_ = json.Unmarshal([]byte(nullStr(hints)), &p.CoachHints)
	if p.Tags == nil {
		p.Tags = []string{}
	}
	if p.CoachHints == nil {
		p.CoachHints = []string{}
	}
	return p, nil
}

func nullStr(ns sql.NullString) string {
	if ns.Valid && ns.String != "" {
		return ns.String
	}
	return "[]"
}
