package proxy

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/golang/mock/gomock"
)

func TestRunSequentialTaskFailOnError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockRemoteClient(ctrl)
	gomock.InOrder(
		m.EXPECT().Update(gomock.Any(), "addr0", "info").Return(nil),
		m.EXPECT().Update(gomock.Any(), "addr1", "info").Return(errors.New("boom")),
	)

	task, err := newTask(&TaskConfig{
		Mode:        Sequential,
		FailOnError: true,
		Info:        "info",
	}, m, []string{"addr0", "addr1", "addr2"}, log.NewNopLogger())
	if err != nil {
		panic(err)
	}

	<-task.context.Done()

	s := task.status()

	if !reflect.DeepEqual(s, &TaskStatus{
		Results: []Result{
			{
				Addr:   "addr0",
				Status: Success,
			},
			{
				Addr:   "addr1",
				Status: Failure,
				Msg:    "boom",
			},
			{
				Addr:   "addr2",
				Status: Ignored,
			},
		},
	}) {
		t.Fatal("wrong status", s)
	}
}

func TestRunSequentialTask(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockRemoteClient(ctrl)
	gomock.InOrder(
		m.EXPECT().Update(gomock.Any(), "addr0", "info").Return(nil),
		m.EXPECT().Update(gomock.Any(), "addr1", "info").Return(errors.New("boom")),
		m.EXPECT().Update(gomock.Any(), "addr2", "info").Return(errors.New("boom")),
	)

	task, err := newTask(&TaskConfig{
		Mode:        Sequential,
		FailOnError: false,
		Info:        "info",
	}, m, []string{"addr0", "addr1", "addr2"}, log.NewNopLogger())
	if err != nil {
		panic(err)
	}

	<-task.done

	s := task.status()

	if !reflect.DeepEqual(s, &TaskStatus{
		Results: []Result{
			{
				Addr:   "addr0",
				Status: Success,
			},
			{
				Addr:   "addr1",
				Status: Failure,
				Msg:    "boom",
			},
			{
				Addr:   "addr2",
				Status: Failure,
				Msg:    "boom",
			},
		},
	}) {
		t.Fatal("wrong status", s)
	}
}

func TestRunSequentialTaskKill(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockRemoteClient(ctrl)
	m.EXPECT().Update(gomock.Any(), "addr0", "info").Return(errors.New("killed")).Do(func(ctx context.Context, addr, info string) { <-ctx.Done() })

	task, err := newTask(&TaskConfig{
		Mode:        Sequential,
		FailOnError: false,
		Info:        "info",
	}, m, []string{"addr0", "addr1", "addr2"}, log.NewNopLogger())
	if err != nil {
		panic(err)
	}

	task.kill()

	<-task.done

	s := task.status()

	if !reflect.DeepEqual(s, &TaskStatus{
		Results: []Result{
			{
				Addr:   "addr0",
				Status: Killed,
			},
			{
				Addr:   "addr1",
				Status: Ignored,
			},
			{
				Addr:   "addr2",
				Status: Ignored,
			},
		},
	}) {
		t.Fatal("wrong status", s)
	}
}

func TestRunParallelTaskFailOnError(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockRemoteClient(ctrl)
	m.EXPECT().Update(gomock.Any(), "addr0", "info").Return(errors.New("killed")).Do(func(ctx context.Context, addr, info string) { <-ctx.Done() })
	m.EXPECT().Update(gomock.Any(), "addr1", "info").Return(errors.New("boom"))
	m.EXPECT().Update(gomock.Any(), "addr2", "info").Return(errors.New("killed")).Do(func(ctx context.Context, addr, info string) { <-ctx.Done() })

	task, err := newTask(&TaskConfig{
		Mode:        Parallel,
		FailOnError: true,
		Info:        "info",
	}, m, []string{"addr0", "addr1", "addr2"}, log.NewNopLogger())
	if err != nil {
		panic(err)
	}

	<-task.done

	s := task.status()

	if !reflect.DeepEqual(s, &TaskStatus{
		Results: []Result{
			{
				Addr:   "addr0",
				Status: Killed,
			},
			{
				Addr:   "addr1",
				Status: Failure,
				Msg:    "boom",
			},
			{
				Addr:   "addr2",
				Status: Killed,
			},
		},
	}) {
		t.Fatal("wrong status", s)
	}
}

func TestRunParallelTask(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockRemoteClient(ctrl)
	m.EXPECT().Update(gomock.Any(), "addr0", "info").Return(nil)
	m.EXPECT().Update(gomock.Any(), "addr1", "info").Return(errors.New("boom"))
	m.EXPECT().Update(gomock.Any(), "addr2", "info").Return(nil)

	task, err := newTask(&TaskConfig{
		Mode:        Parallel,
		FailOnError: false,
		Info:        "info",
	}, m, []string{"addr0", "addr1", "addr2"}, log.NewNopLogger())
	if err != nil {
		panic(err)
	}

	<-task.done

	s := task.status()

	if !reflect.DeepEqual(s, &TaskStatus{
		Results: []Result{
			{
				Addr:   "addr0",
				Status: Success,
			},
			{
				Addr:   "addr1",
				Status: Failure,
				Msg:    "boom",
			},
			{
				Addr:   "addr2",
				Status: Success,
			},
		},
	}) {
		t.Fatal("wrong status", s)
	}
}
