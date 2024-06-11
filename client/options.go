package client

import (
	"errors"
	"net/url"
)

// Option configures a Client
type Option func(*Client) error

// BaseURL is a Client's option to set the baseURL of the REST client.
func BaseURL(URL *url.URL) Option {
	return func(c *Client) error {
		c.baseURL = URL
		return nil
	}
}

func UserAgent(userAgent string) Option {
	return func(c *Client) error {
		c.userAgent = userAgent
		return nil
	}
}

// Log is a client's option to set a Logger
func Log(logger interface{}) Option {
	return func(c *Client) error {
		switch logger.(type) {
		case Logger, TestLogger, LeveledLogger, nil:
			c.log = logger
		default:
			return errors.New("invalid logger type, expected Logger or LeveledLogger")
		}
		return nil
	}
}
