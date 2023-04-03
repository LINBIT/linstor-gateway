package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/LINBIT/linstor-gateway/pkg/rest"
	"github.com/moul/http2curl"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Client struct {
	httpClient *http.Client
	baseURL    *url.URL
	log        interface{} // must be either Logger, TestLogger, or LeveledLogger

	Iscsi  *ISCSIService
	Nfs    *NFSService
	NvmeOf *NvmeOfService
	Status *StatusService
}

type clientError string

func (e clientError) Error() string { return string(e) }

const (
	// NotFoundError is the error type returned in case of a 404 error. This is required to test for this kind of error.
	NotFoundError = clientError("404 Not Found")
)

type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// Logger represents a standard logger interface
type Logger interface {
	Printf(string, ...interface{})
}

// TestLogger represents a logger interface used in tests, as in testing.T
type TestLogger interface {
	Logf(string, ...interface{})
}

// LeveledLogger interface implements the basic methods that a logger library needs
type LeveledLogger interface {
	Errorf(string, ...interface{})
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
	Warnf(string, ...interface{})
}

func NewClient(options ...Option) (*Client, error) {
	defaultBase, err := url.Parse("http://localhost:8080")
	if err != nil {
		return nil, fmt.Errorf("failed to parse default URL: %w", err)
	}

	c := &Client{
		httpClient: &http.Client{},
		log:        log.New(os.Stderr, "", 0),
		baseURL:    defaultBase,
	}

	for _, opt := range options {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	c.Iscsi = &ISCSIService{c}
	c.Nfs = &NFSService{c}
	c.NvmeOf = &NvmeOfService{c}
	c.Status = &StatusService{c}
	return c, nil
}

func (c *Client) logf(level LogLevel, msg string, args ...interface{}) {
	switch l := c.log.(type) {
	case LeveledLogger:
		switch level {
		case LevelDebug:
			l.Debugf(msg, args...)
		case LevelInfo:
			l.Infof(msg, args...)
		case LevelWarn:
			l.Warnf(msg, args...)
		case LevelError:
			l.Errorf(msg, args...)
		}
	case TestLogger:
		l.Logf("[%s] %s", level, fmt.Sprintf(msg, args...))
	case Logger:
		l.Printf("[%s] %s", level, fmt.Sprintf(msg, args...))
	}
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := c.baseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
		c.logf(LevelDebug, "%s", buf)
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Client) curlify(req *http.Request) (string, error) {
	cc, err := http2curl.GetCurlCommand(req)
	if err != nil {
		return "", err
	}
	return cc.String(), nil
}

func (c *Client) logCurlify(req *http.Request) {
	var msg string
	if curl, err := c.curlify(req); err != nil {
		msg = err.Error()
	} else {
		msg = curl
	}

	c.logf(LevelDebug, "%s", msg)
}

func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	req = req.WithContext(ctx)

	c.logCurlify(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		msg := fmt.Sprintf("Status code not within 200 to 400, but %d (%s)",
			resp.StatusCode, http.StatusText(resp.StatusCode))
		c.logf(LevelDebug, "%s", msg)
		if resp.StatusCode == 404 {
			return nil, NotFoundError
		}

		var e rest.Error
		err = json.NewDecoder(resp.Body).Decode(&e)
		if err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, e
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return resp, err
}

func (c *Client) doGET(ctx context.Context, url string, ret interface{}) (*http.Response, error) {
	req, err := c.newRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.do(ctx, req, &ret)
}

func (c *Client) doPOST(ctx context.Context, url string, body interface{}, ret interface{}) (*http.Response, error) {
	req, err := c.newRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	return c.do(ctx, req, &ret)
}

func (c *Client) doPUT(ctx context.Context, url string, body interface{}, ret interface{}) (*http.Response, error) {
	req, err := c.newRequest("PUT", url, body)
	if err != nil {
		return nil, err
	}

	return c.do(ctx, req, &ret)
}

func (c *Client) doDELETE(ctx context.Context, url string, body interface{}) (*http.Response, error) {
	req, err := c.newRequest("DELETE", url, body)
	if err != nil {
		return nil, err
	}

	return c.do(ctx, req, nil)
}
