package db

import (
	"database/sql"
	"fmt"
)

type CounselingBooking struct {
	ID             string `json:"id"`
	UserID         string `json:"userId"`
	PreferredSlots string `json:"preferredSlots"`
	Note           string `json:"note"`
	ContactChannel string `json:"contactChannel"`
	Status         string `json:"status"` // requested | confirmed | done | cancelled
	CreatedAt      string `json:"createdAt"`
	UpdatedAt      string `json:"updatedAt"`
}

func (s *Store) CreateCounselingBooking(b *CounselingBooking) error {
	_, err := s.db.Exec(`
INSERT INTO counseling_bookings (
  id, user_id, preferred_slots, note, contact_channel, status, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		b.ID, b.UserID, b.PreferredSlots, b.Note, b.ContactChannel, b.Status, b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (s *Store) ListCounselingBookings(userID string, limit int) ([]CounselingBooking, error) {
	if limit <= 0 {
		limit = 30
	}
	rows, err := s.db.Query(`
SELECT id, user_id, preferred_slots, note, contact_channel, status, created_at, updated_at
FROM counseling_bookings WHERE user_id = ?
ORDER BY updated_at DESC LIMIT ?`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]CounselingBooking, 0)
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) GetCounselingBooking(id, userID string) (*CounselingBooking, error) {
	row := s.db.QueryRow(`
SELECT id, user_id, preferred_slots, note, contact_channel, status, created_at, updated_at
FROM counseling_bookings WHERE id = ? AND user_id = ?`, id, userID)
	b, err := scanBooking(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Store) UpdateCounselingBooking(b *CounselingBooking) error {
	res, err := s.db.Exec(`
UPDATE counseling_bookings SET
  preferred_slots = ?, note = ?, contact_channel = ?, status = ?, updated_at = ?
WHERE id = ? AND user_id = ?`,
		b.PreferredSlots, b.Note, b.ContactChannel, b.Status, b.UpdatedAt, b.ID, b.UserID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("booking not found")
	}
	return nil
}

func scanBooking(sc rowScanner) (CounselingBooking, error) {
	var b CounselingBooking
	err := sc.Scan(
		&b.ID, &b.UserID, &b.PreferredSlots, &b.Note, &b.ContactChannel, &b.Status, &b.CreatedAt, &b.UpdatedAt,
	)
	return b, err
}
