package proxy

import (
	"context"
	"errors"
	"sync"

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

// NewService creates new service instance.
func NewService(client RemoteClient, addrs []string, logger log.Logger) Service {
	if client == nil {
		panic("missing client")
	}
	if addrs == nil {
		panic("missing addrs")
	}
	if logger == nil {
		panic("missing logger")
	}

	return &service{
		client: client,
		addrs:  addrs,
		tasks:  make(map[TaskID]*task),
		logger: logger,
	}
}

func (s *service) CreateTask(ctx context.Context, config *TaskConfig) (TaskID, error) {
	t, err := newTask(config, s.client, s.addrs, s.logger)
	if err != nil {
		s.logger.Log(
			"msg", "failed to create task",
			"err", err,
		)
		return "", errors.New("failed to generate id")
	}

	s.tasksMu.Lock()
	s.tasks[t.ID()] = t
	s.tasksMu.Unlock()

	return t.ID(), nil
}

func (s *service) TaskStatus(ctx context.Context, id TaskID) (*TaskStatus, error) {
	s.tasksMu.RLock()
	t := s.tasks[id]
	s.tasksMu.RUnlock()

	if t == nil {
		return nil, nil
	}

	return t.status(), nil
}

func (s *service) KillTask(ctx context.Context, id TaskID) (*TaskStatus, error) {
	s.tasksMu.RLock()
	t := s.tasks[id]
	s.tasksMu.RUnlock()

	if t == nil {
		return nil, nil
	}

	t.kill()

	return t.status(), nil
}
