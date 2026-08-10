package db

import (
	"database/sql"
	"fmt"
)

type WellbeingCheckIn struct {
	ID                     string   `json:"id"`
	UserID                 string   `json:"userId"`
	At                     string   `json:"at"`
	StressScore            int      `json:"stressScore"`
	MoodTags               []string `json:"moodTags"`
	EnergyScore            int      `json:"energyScore"`
	EventType              string   `json:"eventType"`
	Note                   string   `json:"note"`
	RelatedTaskID          string   `json:"relatedTaskId,omitempty"`
	RelatedCoachSessionID  string   `json:"relatedCoachSessionId,omitempty"`
	CreatedAt              string   `json:"createdAt"`
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
