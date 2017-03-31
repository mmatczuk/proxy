package proxy

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RemoteClient provides ability to call the legacy system, implementations must
// be thread safe.
type RemoteClient interface {
	Update(ctx context.Context, addr, info string) error
}

type remoteClient struct {
	url    url.URL
	client http.Client
}

func NewRemoteClient() *remoteClient {
	return &remoteClient{
		client: http.Client{
			Transport: &http.Transport{
				Dial: func(network, addr string) (conn net.Conn, err error) {
					conn, err = net.Dial(network, addr)
					if conn != nil {
						err = conn.SetDeadline(time.Time{})
					}
					return
				},
			},
		},
	}
}

func (c *remoteClient) Update(ctx context.Context, addr, info string) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s", addr), strings.NewReader(info))
	if err != nil {
		return fmt.Errorf("failed to create request: %s", err)
	}
	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %s", err)
	}

	b, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		return fmt.Errorf("failed to read response: %s", err)
	}

	if string(b[0:2]) != "OK" {
		return fmt.Errorf("remote failure: %s", b)
	}

	return nil
}
