package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type WellbeingCheckIn struct {
	ID                    string   `json:"id"`
	UserID                string   `json:"userId"`
	At                    string   `json:"at"`
	StressScore           int      `json:"stressScore"`
	MoodTags              []string `json:"moodTags"`
	EnergyScore           int      `json:"energyScore"`
	EventType             string   `json:"eventType"`
	Note                  string   `json:"note"`
	RelatedTaskID         string   `json:"relatedTaskId,omitempty"`
	RelatedCoachSessionID string   `json:"relatedCoachSessionId,omitempty"`
	CreatedAt             string   `json:"createdAt"`
}

// QuickSelfCheck is the 3-minute workplace self-check (non-clinical).
type QuickSelfCheck struct {
	ID                    string   `json:"id"`
	UserID                string   `json:"userId"`
	At                    string   `json:"at"`
	Version               string   `json:"version"`
	Feelings              []string `json:"feelings"`
	Duration              string   `json:"duration"`
	Impacts               []string `json:"impacts"`
	DistressScore         int      `json:"distressScore"`
	TriggerNote           string   `json:"triggerNote,omitempty"`
	Takeaway              string   `json:"takeaway"`
	RelatedCoachSessionID string   `json:"relatedCoachSessionId,omitempty"`
	CreatedAt             string   `json:"createdAt"`
}

func (s *Store) CreateCheckIn(c *WellbeingCheckIn) error {
	_, err := s.db.Exec(`
INSERT INTO wellbeing_checkins (
  id, user_id, at, stress_score, mood_tags, energy_score, event_type, note,
  related_task_id, related_coach_session_id, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.UserID, c.At, c.StressScore, mustJSON(c.MoodTags), c.EnergyScore, c.EventType, c.Note,
		c.RelatedTaskID, c.RelatedCoachSessionID, c.CreatedAt,
	)
	return err
}

func (s *Store) ListCheckIns(userID string, limit int) ([]WellbeingCheckIn, error) {
	if limit <= 0 {
		limit = 90
	}
	rows, err := s.db.Query(`
SELECT id, user_id, at, stress_score, mood_tags, energy_score, event_type, note,
       related_task_id, related_coach_session_id, created_at
FROM wellbeing_checkins WHERE user_id = ?
ORDER BY at DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]WellbeingCheckIn, 0)
	for rows.Next() {
		c, err := scanCheckIn(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) DeleteCheckIn(id, userID string) error {
	res, err := s.db.Exec(`DELETE FROM wellbeing_checkins WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("checkin not found")
	}
	return nil
}

func scanCheckIn(sc rowScanner) (WellbeingCheckIn, error) {
	var c WellbeingCheckIn
	var moods sql.NullString
	err := sc.Scan(
		&c.ID, &c.UserID, &c.At, &c.StressScore, &moods, &c.EnergyScore, &c.EventType, &c.Note,
		&c.RelatedTaskID, &c.RelatedCoachSessionID, &c.CreatedAt,
	)
	if err != nil {
		return WellbeingCheckIn{}, err
	}
	c.MoodTags = decodeStringSlice(moods)
	if c.MoodTags == nil {
		c.MoodTags = []string{}
	}
	return c, nil
}

func (s *Store) CreateQuickSelfCheck(c *QuickSelfCheck) error {
	_, err := s.db.Exec(`
INSERT INTO quick_self_checks (
  id, user_id, at, version, feelings, duration, impacts, distress_score,
  trigger_note, takeaway, related_coach_session_id, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID, c.UserID, c.At, c.Version, mustJSON(c.Feelings), c.Duration, mustJSON(c.Impacts),
		c.DistressScore, c.TriggerNote, c.Takeaway, c.RelatedCoachSessionID, c.CreatedAt,
	)
	return err
}

func (s *Store) ListQuickSelfChecks(userID string, limit int) ([]QuickSelfCheck, error) {
	if limit <= 0 {
		limit = 30
	}
	rows, err := s.db.Query(`
SELECT id, user_id, at, version, feelings, duration, impacts, distress_score,
       trigger_note, takeaway, related_coach_session_id, created_at
FROM quick_self_checks WHERE user_id = ?
ORDER BY at DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]QuickSelfCheck, 0)
	for rows.Next() {
		c, err := scanQuickSelfCheck(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) LatestQuickSelfCheckSince(userID string, since time.Time) (*QuickSelfCheck, error) {
	items, err := s.ListQuickSelfChecks(userID, 5)
	if err != nil {
		return nil, err
	}
	for i := range items {
		t, err := time.Parse(time.RFC3339, items[i].At)
		if err != nil {
			continue
		}
		if !t.Before(since) {
			return &items[i], nil
		}
	}
	return nil, nil
}

func (s *Store) LatestQuickSelfCheck(userID string) (*QuickSelfCheck, error) {
	items, err := s.ListQuickSelfChecks(userID, 1)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	return &items[0], nil
}

func (s *Store) QuickSelfCheckByCoachSession(userID, sessionID string) (*QuickSelfCheck, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, nil
	}
	row := s.db.QueryRow(`
SELECT id, user_id, at, version, feelings, duration, impacts, distress_score,
       trigger_note, takeaway, related_coach_session_id, created_at
FROM quick_self_checks
WHERE user_id = ? AND related_coach_session_id = ?
ORDER BY created_at DESC LIMIT 1`, userID, sessionID)
	c, err := scanQuickSelfCheck(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) GetQuickSelfCheck(id, userID string) (*QuickSelfCheck, error) {
	row := s.db.QueryRow(`
SELECT id, user_id, at, version, feelings, duration, impacts, distress_score,
       trigger_note, takeaway, related_coach_session_id, created_at
FROM quick_self_checks WHERE id = ? AND user_id = ?`, id, userID)
	c, err := scanQuickSelfCheck(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) UpdateQuickSelfCheck(c *QuickSelfCheck) error {
	res, err := s.db.Exec(`
UPDATE quick_self_checks SET
  related_coach_session_id = ?
WHERE id = ? AND user_id = ?`,
		c.RelatedCoachSessionID, c.ID, c.UserID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("quick self-check not found")
	}
	return nil
}

func (s *Store) DeleteQuickSelfCheck(id, userID string) error {
	res, err := s.db.Exec(`DELETE FROM quick_self_checks WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("quick self-check not found")
	}
	return nil
}

func scanQuickSelfCheck(sc rowScanner) (QuickSelfCheck, error) {
	var c QuickSelfCheck
	var feelings, impacts sql.NullString
	err := sc.Scan(
		&c.ID, &c.UserID, &c.At, &c.Version, &feelings, &c.Duration, &impacts, &c.DistressScore,
		&c.TriggerNote, &c.Takeaway, &c.RelatedCoachSessionID, &c.CreatedAt,
	)
	if err != nil {
		return QuickSelfCheck{}, err
	}
	c.Feelings = decodeStringSlice(feelings)
	c.Impacts = decodeStringSlice(impacts)
	if c.Feelings == nil {
		c.Feelings = []string{}
	}
	if c.Impacts == nil {
		c.Impacts = []string{}
	}
	return c, nil
}
