package proxy

import "context"

type task struct {
	context context.Context
	cancel  context.CancelFunc
	config  TaskConfig
}

func newTask(config *TaskConfig, client RemoteClient, addrs ...string) (*task, error) {
	return nil, nil
}

func (t *task) status() (*TaskStatus, error) {
	return nil, nil
}

func (t *task) kill() {
	t.cancel()
	t.context.Err()
}
