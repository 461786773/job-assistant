package db

import (
	"database/sql"
	"encoding/json"
)

type InitialAssessment struct {
	ID                 string          `json:"id"`
	UserID             string          `json:"userId"`
	Version            string          `json:"version"`
	Answers            json.RawMessage `json:"answers"`
	PrimaryScene       string          `json:"primaryScene"`
	WorkStatus         string          `json:"workStatus"`
	KeyEvents          []string        `json:"keyEvents"`
	TenureBand         string          `json:"tenureBand"`
	Scores             json.RawMessage `json:"scores"`
	MoodTags           []string        `json:"moodTags"`
	Stressors          []string        `json:"stressors"`
	Coping             []string        `json:"coping"`
	SupportLevel       string          `json:"supportLevel"`
	CheckinWillingness string          `json:"checkinWillingness"`
	CrisisLevel        string          `json:"crisisLevel"` // none | fleeting | elevated
	Goals              []string        `json:"goals"`
	FreeTextBlockers   string          `json:"freeTextBlockers,omitempty"`
	FreeTextOther      string          `json:"freeTextOther,omitempty"`
	Metrics            json.RawMessage `json:"metrics,omitempty"`
	AIAnalysis         json.RawMessage `json:"aiAnalysis,omitempty"`
	SummaryForCoach    string          `json:"summaryForCoach,omitempty"`
	CompletedAt        string          `json:"completedAt"`
	CreatedAt          string          `json:"createdAt"`
}

func (s *Store) migrateAssessmentBooking() error {
	if err := s.ensureColumn("users", "primary_need", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
	}
	const schema = `
CREATE TABLE IF NOT EXISTS initial_assessments (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  version TEXT NOT NULL DEFAULT 'v1',
  answers TEXT,
  primary_scene TEXT NOT NULL DEFAULT '',
  work_status TEXT NOT NULL DEFAULT '',
  key_events TEXT,
  tenure_band TEXT NOT NULL DEFAULT '',
  scores TEXT,
  mood_tags TEXT,
  stressors TEXT,
  coping TEXT,
  support_level TEXT NOT NULL DEFAULT '',
  checkin_willingness TEXT NOT NULL DEFAULT '',
  crisis_level TEXT NOT NULL DEFAULT 'none',
  goals TEXT,
  free_text_blockers TEXT NOT NULL DEFAULT '',
  free_text_other TEXT NOT NULL DEFAULT '',
  metrics TEXT,
  ai_analysis TEXT,
  summary_for_coach TEXT NOT NULL DEFAULT '',
  completed_at TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_assessments_user_completed ON initial_assessments(user_id, completed_at);
CREATE TABLE IF NOT EXISTS counseling_bookings (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  preferred_slots TEXT NOT NULL DEFAULT '',
  note TEXT NOT NULL DEFAULT '',
  contact_channel TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'requested',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_bookings_user_updated ON counseling_bookings(user_id, updated_at);
`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) CreateInitialAssessment(a *InitialAssessment) error {
	_, err := s.db.Exec(`
INSERT INTO initial_assessments (
  id, user_id, version, answers, primary_scene, work_status, key_events, tenure_band,
  scores, mood_tags, stressors, coping, support_level, checkin_willingness, crisis_level,
  goals, free_text_blockers, free_text_other, metrics, ai_analysis, summary_for_coach,
  completed_at, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.UserID, a.Version, nullJSON(a.Answers), a.PrimaryScene, a.WorkStatus, mustJSON(a.KeyEvents), a.TenureBand,
		nullJSON(a.Scores), mustJSON(a.MoodTags), mustJSON(a.Stressors), mustJSON(a.Coping), a.SupportLevel, a.CheckinWillingness, a.CrisisLevel,
		mustJSON(a.Goals), a.FreeTextBlockers, a.FreeTextOther, nullJSON(a.Metrics), nullJSON(a.AIAnalysis), a.SummaryForCoach,
		a.CompletedAt, a.CreatedAt,
	)
	return err
}

func (s *Store) ListInitialAssessments(userID string, limit int) ([]InitialAssessment, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(`
SELECT id, user_id, version, answers, primary_scene, work_status, key_events, tenure_band,
       scores, mood_tags, stressors, coping, support_level, checkin_willingness, crisis_level,
       goals, free_text_blockers, free_text_other, metrics, ai_analysis, summary_for_coach,
       completed_at, created_at
FROM initial_assessments WHERE user_id = ?
ORDER BY completed_at DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]InitialAssessment, 0)
	for rows.Next() {
		a, err := scanInitialAssessment(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) GetInitialAssessment(id, userID string) (*InitialAssessment, error) {
	row := s.db.QueryRow(`
SELECT id, user_id, version, answers, primary_scene, work_status, key_events, tenure_band,
       scores, mood_tags, stressors, coping, support_level, checkin_willingness, crisis_level,
       goals, free_text_blockers, free_text_other, metrics, ai_analysis, summary_for_coach,
       completed_at, created_at
FROM initial_assessments WHERE id = ? AND user_id = ?`, id, userID)
	a, err := scanInitialAssessment(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Store) LatestInitialAssessment(userID string) (*InitialAssessment, error) {
	row := s.db.QueryRow(`
SELECT id, user_id, version, answers, primary_scene, work_status, key_events, tenure_band,
       scores, mood_tags, stressors, coping, support_level, checkin_willingness, crisis_level,
       goals, free_text_blockers, free_text_other, metrics, ai_analysis, summary_for_coach,
       completed_at, created_at
FROM initial_assessments WHERE user_id = ?
ORDER BY completed_at DESC LIMIT 1`, userID)
	a, err := scanInitialAssessment(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Store) HasInitialAssessment(userID string) (bool, error) {
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM initial_assessments WHERE user_id = ?`, userID).Scan(&n)
	return n > 0, err
}

func scanInitialAssessment(sc rowScanner) (InitialAssessment, error) {
	var a InitialAssessment
	var answers, keyEvents, scores, moods, stressors, coping, goals, metrics, ai sql.NullString
	err := sc.Scan(
		&a.ID, &a.UserID, &a.Version, &answers, &a.PrimaryScene, &a.WorkStatus, &keyEvents, &a.TenureBand,
		&scores, &moods, &stressors, &coping, &a.SupportLevel, &a.CheckinWillingness, &a.CrisisLevel,
		&goals, &a.FreeTextBlockers, &a.FreeTextOther, &metrics, &ai, &a.SummaryForCoach,
		&a.CompletedAt, &a.CreatedAt,
	)
	if err != nil {
		return InitialAssessment{}, err
	}
	a.Answers = rawJSON(answers)
	a.Scores = rawJSON(scores)
	a.Metrics = rawJSON(metrics)
	a.AIAnalysis = rawJSON(ai)
	a.KeyEvents = decodeStringSlice(keyEvents)
	a.MoodTags = decodeStringSlice(moods)
	a.Stressors = decodeStringSlice(stressors)
	a.Coping = decodeStringSlice(coping)
	a.Goals = decodeStringSlice(goals)
	if a.KeyEvents == nil {
		a.KeyEvents = []string{}
	}
	if a.MoodTags == nil {
		a.MoodTags = []string{}
	}
	if a.Stressors == nil {
		a.Stressors = []string{}
	}
	if a.Coping == nil {
		a.Coping = []string{}
	}
	if a.Goals == nil {
		a.Goals = []string{}
	}
	return a, nil
}
