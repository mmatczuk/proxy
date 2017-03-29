package proxy

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/mmatczuk/proxy/log"
)

// Service provides proxy operations.
type Service interface {
	CreateTask(ctx context.Context, config *TaskConfig) (TaskID, error)
	TaskStatus(ctx context.Context, id TaskID) (*TaskStatus, error)
	KillTask(ctx context.Context, id TaskID) (*TaskStatus, error)
}

type service struct {
	client  RemoteClient
	addrs   []string
	tasks   map[TaskID]*task
	tasksMu sync.RWMutex
	logger  log.Logger
}

func (s *service) CreateTask(ctx context.Context, config *TaskConfig) (TaskID, error) {
	id, err := newTaskID()
	if err != nil {
		s.logger.Log(
			"msg", "failed to generate id",
			"err", err,
		)
		return "", errors.New("failed to generate id")
	}

	t, err := newTask(config, s.client, s.addrs...)
	if err != nil {
		s.logger.Log(
			"msg", "failed to create task",
			"err", err,
		)
		return "", errors.New("failed to create task")
	}

	s.tasksMu.Lock()
	s.tasks[id] = t
	s.tasksMu.Unlock()

	return id, nil
}

func newTaskID() (TaskID, error) {
	u, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}
	return TaskID(u.String()), nil
}

func (s *service) TaskStatus(ctx context.Context, id TaskID) (*TaskStatus, error) {
	s.tasksMu.RLock()
	t := s.tasks[id]
	s.tasksMu.RUnlock()

	if t == nil {
		return nil, nil
	}

	return s.status(t)
}

func (s *service) KillTask(ctx context.Context, id TaskID) (*TaskStatus, error) {
	s.tasksMu.RLock()
	t := s.tasks[id]
	s.tasksMu.RUnlock()

	if t == nil {
		return nil, nil
	}

	t.kill()

	return s.status(t)
}

func (s *service) status(t *task) (*TaskStatus, error) {
	status, err := t.status()
	if err != nil {
		s.logger.Log(
			"msg", "failed to get status",
			"err", err,
		)
		return nil, errors.New("failed to get status")
	}

	return status, nil
}
