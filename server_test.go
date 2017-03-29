package proxy

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestServerCreateTask(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockService(ctrl)
	m.EXPECT().CreateTask(gomock.Any(), &TaskConfig{
		ClientID:    "f0a4fd40-44bf-4535-b807-632586645d6f",
		Info:        "test",
		Mode:        Sequential,
		FailOnError: true,
	}).Return(TaskID("test"), nil)
	s := NewServer(m)

	body := `{
  "client_id": "f0a4fd40-44bf-4535-b807-632586645d6f",
  "info": "test",
  "mode": "sequential",
  "failonerror": true
}
`

	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/v1/task", strings.NewReader(body)))

	if w.Code != http.StatusCreated {
		t.Fatal("wrong status code", w)
	}
	if strings.TrimSpace(w.Body.String()) != `"test"` {
		t.Fatal("wrong body", w)
	}
}

func TestServerCreateTaskError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockService(ctrl)
	m.EXPECT().CreateTask(gomock.Any(), gomock.Any()).Return(TaskID(""), errors.New("foobar"))
	s := NewServer(m)

	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/v1/task", strings.NewReader("{}")))

	if w.Code != http.StatusInternalServerError {
		t.Fatal("wrong status code", w)
	}
	if strings.TrimSpace(w.Body.String()) != "foobar" {
		t.Fatal("wrong body", w)
	}
}

func TestSeverTaskStatus(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockService(ctrl)
	m.EXPECT().TaskStatus(gomock.Any(), TaskID("test")).Return(&TaskStatus{
		Results: []Result{
			{
				Addr:   "addr:1",
				Status: Success,
			},
			{
				Addr:   "addr:2",
				Status: Failure,
				Msg:    "foobar",
			},
		},
	}, nil)
	s := NewServer(m)

	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/task/test/status", nil))

	if w.Code != http.StatusOK {
		t.Fatal("wrong status code", w)
	}

	if strings.TrimSpace(w.Body.String()) != `[{"addr":"addr:1","status":"success"},{"addr":"addr:2","status":"failure","message":"foobar"}]` {
		t.Fatal("wrong body", w)
	}
}

func TestSeverTaskStatusError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockService(ctrl)
	m.EXPECT().TaskStatus(gomock.Any(), TaskID("test")).Return(nil, errors.New("foobar"))
	s := NewServer(m)

	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/task/test/status", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatal("wrong status code", w)
	}
	if strings.TrimSpace(w.Body.String()) != "foobar" {
		t.Fatal("wrong body", w)
	}
}

func TestSeverKillTask(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockService(ctrl)
	m.EXPECT().KillTask(gomock.Any(), TaskID("test")).Return(&TaskStatus{
		Results: []Result{
			{
				Addr:   "addr",
				Status: Running,
			},
			{
				Addr:   "addr:1",
				Status: Success,
			},
			{
				Addr:   "addr:2",
				Status: Failure,
			},
			{
				Addr:   "addr:3",
				Status: Killed,
			},
		},
	}, nil)
	s := NewServer(m)

	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/task/test/kill", nil))

	if w.Code != http.StatusOK {
		t.Fatal("wrong status code", w)
	}

	if strings.TrimSpace(w.Body.String()) != `[{"addr":"addr:3","status":"killed"}]` {
		t.Fatal("wrong body", w)
	}
}

func TestSeverKillTaskError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockService(ctrl)
	m.EXPECT().KillTask(gomock.Any(), TaskID("test")).Return(nil, errors.New("foobar"))
	s := NewServer(m)

	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/task/test/kill", nil))

	if w.Code != http.StatusInternalServerError {
		t.Fatal("wrong status code", w)
	}
	if strings.TrimSpace(w.Body.String()) != "foobar" {
		t.Fatal("wrong body", w)
	}
}
