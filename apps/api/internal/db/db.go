package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Store struct {
	path  string
	mu    sync.Mutex
	tasks map[string]Task
}

type Task struct {
	ID             string          `json:"id"`
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
	s := &Store{
		path:  path,
		tasks: map[string]Task{},
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return nil
}

func (s *Store) Migrate() error {
	return s.persist()
}

func Now() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
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
		return fmt.Errorf("parse store: %w", err)
	}
	for _, t := range disk.Tasks {
		s.tasks[t.ID] = t
	}
	return nil
}

func (s *Store) persist() error {
	disk := diskFormat{Tasks: make([]Task, 0, len(s.tasks))}
	for _, t := range s.tasks {
		disk.Tasks = append(disk.Tasks, t)
	}
	// newest first for stable-ish listing before sort in ListTasks
	data, err := json.MarshalIndent(disk, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) ListTasks() ([]Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		out = append(out, t)
	}
	// sort by UpdatedAt desc
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].UpdatedAt > out[i].UpdatedAt {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out, nil
}

func (s *Store) GetTask(id string) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, nil
	}
	cp := t
	return &cp, nil
}

func (s *Store) CreateTask(t *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[t.ID] = *t
	return s.persist()
}

func (s *Store) UpdateTask(t *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[t.ID]; !ok {
		return fmt.Errorf("task not found")
	}
	s.tasks[t.ID] = *t
	return s.persist()
}

func (s *Store) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
	return s.persist()
}
