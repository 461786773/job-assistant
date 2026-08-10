package db

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSQLiteCRUD(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tasks.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer store.Close()

	now := Now()
	task := &Task{
		ID: "t1", UserID: "u1", Title: "Demo", Company: "Acme", TargetRole: "PM",
		JDText: "jd", ResumeText: "resume", Status: "draft",
		HrReport:  json.RawMessage(`{"score":80}`),
		CreatedAt: now, UpdatedAt: now,
	}
	if err := store.CreateTask(task); err != nil {
		t.Fatalf("create: %v", err)
	}

	got, err := store.GetTask("t1", "u1")
	if err != nil || got == nil {
		t.Fatalf("get: %v %#v", err, got)
	}
	if got.Title != "Demo" || string(got.HrReport) != `{"score":80}` {
		t.Fatalf("unexpected task: %#v", got)
	}

	// other user cannot see it
	other, err := store.GetTask("t1", "u2")
	if err != nil || other != nil {
		t.Fatalf("expected isolation, got %#v err=%v", other, err)
	}

	got.Title = "Updated"
	got.UpdatedAt = Now()
	if err := store.UpdateTask(got); err != nil {
		t.Fatalf("update: %v", err)
	}
	list, err := store.ListTasks("u1")
	if err != nil || len(list) != 1 || list[0].Title != "Updated" {
		t.Fatalf("list: %v %#v", err, list)
	}
	empty, err := store.ListTasks("u2")
	if err != nil || len(empty) != 0 {
		t.Fatalf("expected empty list for other user: %v %#v", err, empty)
	}

	if err := store.DeleteTask("t1", "u1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	gone, err := store.GetTask("t1", "u1")
	if err != nil || gone != nil {
		t.Fatalf("expected nil after delete, got %#v err=%v", gone, err)
	}
}

func TestImportJSON(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "tasks.json")
	payload := `{"tasks":[{"id":"j1","title":"FromJSON","company":"","targetRole":"","jdText":"","resumeText":"","resumeFilename":"","resumeFormat":"","status":"draft","notes":"","createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-02T00:00:00Z"}]}`
	if err := os.WriteFile(jsonPath, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}
	store, err := Open(filepath.Join(dir, "tasks.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer store.Close()

	got, err := store.GetTask("j1", "")
	if err != nil || got == nil || got.Title != "FromJSON" {
		t.Fatalf("import failed: %v %#v", err, got)
	}
	if _, err := os.Stat(jsonPath); !os.IsNotExist(err) {
		t.Fatalf("expected tasks.json renamed away, stat err=%v", err)
	}
	if _, err := os.Stat(jsonPath + ".migrated"); err != nil {
		t.Fatalf("expected .migrated backup: %v", err)
	}
}

func TestUserCRUD(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "tasks.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer store.Close()

	def, err := store.GetUserByUsername(DefaultUsername)
	if err != nil || def == nil || def.ID != DefaultUserID {
		t.Fatalf("default user missing: %v %#v", err, def)
	}

	u, err := store.EnsureUser("alice")
	if err != nil || u == nil || u.Username != "alice" {
		t.Fatalf("ensure: %v %#v", err, u)
	}
	again, err := store.EnsureUser("alice")
	if err != nil || again == nil || again.ID != u.ID {
		t.Fatalf("ensure idempotent: %v %#v", err, again)
	}
}

func TestQuickSelfCheckCRUD(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "tasks.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer store.Close()

	now := Now()
	c := &QuickSelfCheck{
		ID: "q1", UserID: "u1", At: now, Version: "v1",
		Feelings: []string{"tired", "afraid"}, Duration: "one_two_weeks",
		Impacts: []string{"sleep", "focus"}, DistressScore: 7,
		TriggerNote: "二面挂了", Takeaway: "strength", CreatedAt: now,
	}
	if err := store.CreateQuickSelfCheck(c); err != nil {
		t.Fatalf("create: %v", err)
	}
	got, err := store.GetQuickSelfCheck("q1", "u1")
	if err != nil || got == nil || got.DistressScore != 7 || len(got.Feelings) != 2 {
		t.Fatalf("get: %v %#v", err, got)
	}
	other, err := store.GetQuickSelfCheck("q1", "u2")
	if err != nil || other != nil {
		t.Fatalf("expected isolation, got %#v err=%v", other, err)
	}
	list, err := store.ListQuickSelfChecks("u1", 10)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v %#v", err, list)
	}
	got.RelatedCoachSessionID = "sess1"
	if err := store.UpdateQuickSelfCheck(got); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := store.DeleteQuickSelfCheck("q1", "u1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
