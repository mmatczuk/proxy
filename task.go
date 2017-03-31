package proxy

import (
	"context"
	"sort"
	"sync"

	"github.com/google/uuid"
	"github.com/mmatczuk/proxy/log"
)

type task struct {
	// ID is task identifier.
	ID TaskID
	// context is a common context for all remote calls.
	context context.Context
	// cancel enables cancelling remote calls.
	cancel context.CancelFunc
	// client performs synchronous remote calls.
	client RemoteClient
	// result maps remote call address to result.
	result map[string]*Result
	// mu protects the task
	mu sync.RWMutex
	// done is closed when task is done
	done chan struct{}
	// logger
	logger log.Logger
}

// newTask creates and starts new asynchronous task.
func newTask(config *TaskConfig, client RemoteClient, addrs []string, logger log.Logger) (*task, error) {
	u, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	t := &task{
		ID:     TaskID(u.String()),
		client: client,
		result: make(map[string]*Result, len(addrs)),
		done:   make(chan struct{}),
		logger: logger,
	}
	t.context, t.cancel = context.WithCancel(context.Background())

	for _, addr := range addrs {
		t.result[addr] = &Result{
			Addr:   addr,
			Status: Pending,
		}
	}

	switch config.Mode {
	case Sequential:
		go t.runSequential(config, addrs)
	case Parallel:
		go t.runParallel(config, addrs)
	default:
		panic("not supported mode")
	}

	return t, nil
}

func (t *task) runSequential(config *TaskConfig, addrs []string) {
	defer t.cancel()
	defer close(t.done)

	for _, addr := range addrs {
		t.remoteCall(config, addr)
		// task was killed
		if t.context.Err() != nil {
			t.markAsKilled()
			break
		}
	}
}

func (t *task) markAsKilled() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, r := range t.result {
		if r.Status == Pending || r.Status == Running {
			r.Status = Killed
		}
	}
}

func (t *task) runParallel(config *TaskConfig, addrs []string) {
	defer t.cancel()
	defer close(t.done)

	var wg sync.WaitGroup
	for _, addr := range addrs {
		addr := addr
		wg.Add(1)
		go func() {
			t.remoteCall(config, addr)
			wg.Done()
		}()
	}
	wg.Wait()
}

func (t *task) remoteCall(config *TaskConfig, addr string) {
	t.setStatus(addr, Running, nil)

	err := t.client.Update(t.context, addr, config.Info)
	if err != nil {
		if t.context.Err() != nil {
			t.setStatus(addr, Killed, nil)
		} else {
			t.setStatus(addr, Failure, err)
		}

		if config.FailOnError {
			t.cancel()
		}

		t.logger.Log(
			"msg", "remote call failure",
			"task", t.ID,
			"addr", addr,
			"err", err,
		)

		return
	}

	t.setStatus(addr, Success, nil)

	t.logger.Log(
		"msg", "remote call success",
		"task", t.ID,
		"addr", addr,
	)
}

func (t *task) status() *TaskStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()

	s := TaskStatus{
		Results: make([]Result, len(t.result), len(t.result)),
	}

	i := 0
	for _, r := range t.result {
		s.Results[i] = *r
		i++
	}

	sort.Slice(s.Results, func(i, j int) bool {
		return s.Results[i].Addr < s.Results[j].Addr
	})

	return &s
}

func (t *task) setStatus(addr string, s Status, err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	r := t.result[addr]
	r.Status = s
	if err != nil {
		r.Msg = err.Error()
	}
}

func (t *task) kill() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cancel()
}
