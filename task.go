package proxy

import (
	"context"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/mmatczuk/proxy/log"
)

// result extends Result with a mutex to protect it's state.
type result struct {
	Result
	// mu protects result
	mu sync.RWMutex
}

func (r *result) setStatus(s Status, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Status = s
	if err != nil {
		r.Msg = err.Error()
	}
}

// task runs remote tasks and stores the results.
type task struct {
	// id is task identifier.
	id TaskID
	// context is a common context for all remote calls.
	context context.Context
	// cancel enables cancelling remote calls.
	cancel context.CancelFunc
	// client performs synchronous remote calls.
	client RemoteClient
	// results contains remote call results.
	results []*result
	// done is closed when task is done
	done chan struct{}
	// logger
	logger log.Logger
}

// newTask creates new task and calls remote systems based on configuration.
func newTask(config *TaskConfig, client RemoteClient, addrs []string, logger log.Logger) (*task, error) {
	u, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	t := &task{
		id:      TaskID(u.String()),
		client:  client,
		results: make([]*result, len(addrs), len(addrs)),
		done:    make(chan struct{}),
		logger:  logger,
	}
	t.context, t.cancel = context.WithCancel(context.Background())

	for i, addr := range addrs {
		t.results[i] = &result{
			Result: Result{
				Addr:   addr,
				Status: Pending,
			},
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

	for i, addr := range addrs {
		err := t.remoteCall(config, addr, t.results[i])
		if t.killed() || (err != nil && config.FailOnError) {
			t.markPendingIgnored()
			break
		}
	}
}

func (t *task) markPendingIgnored() {
	for _, r := range t.results {
		r.mu.Lock()
		if r.Status == Pending {
			r.Status = Ignored
		}
		r.mu.Unlock()
	}
}

func (t *task) runParallel(config *TaskConfig, addrs []string) {
	defer t.cancel()
	defer close(t.done)

	var wg sync.WaitGroup
	for i, addr := range addrs {
		i, addr := i, addr
		wg.Add(1)
		go func() {
			t.remoteCall(config, addr, t.results[i])
			wg.Done()
		}()
	}
	wg.Wait()
}

func (t *task) remoteCall(config *TaskConfig, addr string, r *result) error {
	r.setStatus(Running, nil)

	err := t.client.Update(t.context, addr, config.Info)
	if err != nil {
		if contextCanceledError(err) {
			r.setStatus(Killed, nil)
		} else {
			r.setStatus(Failure, err)
		}

		if config.FailOnError {
			t.cancel()
		}

		t.logger.Log(
			"msg", "remote call failure",
			"task", t.id,
			"addr", addr,
			"err", err,
		)

		return err
	}

	r.setStatus(Success, nil)

	t.logger.Log(
		"msg", "remote call success",
		"task", t.id,
		"addr", addr,
	)

	return nil
}

func contextCanceledError(err error) bool {
	return strings.Contains(err.Error(), context.Canceled.Error())
}

// ID returns taks identifier.
func (t *task) ID() TaskID {
	return t.id
}

func (t *task) status() *TaskStatus {
	s := TaskStatus{
		Results: make([]Result, len(t.results), len(t.results)),
	}

	for i, r := range t.results {
		r.mu.RLock()
		s.Results[i] = r.Result
		r.mu.RUnlock()
	}

	return &s
}

func (t *task) killed() bool {
	return t.context.Err() != nil
}

func (t *task) kill() {
	t.cancel()
	<-t.done
}
