package store

import (
	"errors"
	"strings"
	"sync"

	"practice-2/internal/models"
)

var (
	ErrNotFound     = errors.New("task not found")
	ErrBadTitle     = errors.New("invalid title")
	ErrTitleTooLong = errors.New("title too long")
)

type Store struct {
	mu     sync.Mutex
	nextID int
	tasks  map[int]models.Task
}

func New() *Store {
	return &Store{
		nextID: 1,
		tasks:  make(map[int]models.Task),
	}
}

func (s *Store) Create(title string) (models.Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return models.Task{}, ErrBadTitle
	}
	if len(title) > 120 {
		return models.Task{}, ErrTitleTooLong
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	t := models.Task{
		ID:    s.nextID,
		Title: title,
		Done:  false,
	}
	s.tasks[t.ID] = t
	s.nextID++
	return t, nil
}

func (s *Store) Get(id int) (models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tasks[id]
	if !ok {
		return models.Task{}, ErrNotFound
	}
	return t, nil
}

func (s *Store) List(doneFilter *bool) []models.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]models.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		if doneFilter != nil && t.Done != *doneFilter {
			continue
		}
		out = append(out, t)
	}
	return out
}

func (s *Store) UpdateDone(id int, done bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, ok := s.tasks[id]
	if !ok {
		return ErrNotFound
	}
	t.Done = done
	s.tasks[id] = t
	return nil
}

func (s *Store) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(s.tasks, id)
	return nil
}
