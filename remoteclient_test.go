package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRemoteClientOK(t *testing.T) {
	t.Parallel()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer s.Close()

	c := NewRemoteClient()
	addr := s.Listener.Addr().String()

	err := c.Update(context.Background(), addr, "test")
	if err != nil {
		t.Fatal(err)
	}
}

func TestRemoteClientError(t *testing.T) {
	t.Parallel()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("not quite OK"))
		if err != nil {
			t.Fatal(err)
		}
	}))
	defer s.Close()

	c := NewRemoteClient()
	addr := s.Listener.Addr().String()

	err := c.Update(context.Background(), addr, "test")
	if err == nil || err.Error() != "remote failure: not quite OK" {
		t.Fatal(err)
	}
}

func TestRemoteClientCancel(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {}
	}))
	defer s.Close()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		defer close(done)

		c := NewRemoteClient()
		addr := s.Listener.Addr().String()

		err := c.Update(ctx, addr, "test")
		if err == nil {
			t.Fatal(err)
		}
	}()

	cancel()
	<-done
}
