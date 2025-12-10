package client

import (
	"context"
	"net/http"

	"notifier/log"
	"notifier/log/tag"
)

type DefaultHTTPClient struct {
	c *http.Client
}

func NewDefaultHTTPClient(c *http.Client) HTTPClient {
	return &DefaultHTTPClient{c: c}
}

func (r *DefaultHTTPClient) Do(
	ctx context.Context,
	req *http.Request,
	opts Options,
) (*http.Response, error) {
	resp, err := r.c.Do(req.WithContext(ctx))
	if err != nil {
		log.Error(ctx, "Failed to execute request", tag.Err, err)

		return nil, err
	}

	if opts.ErrHandler == nil {
		opts.ErrHandler = DefaultErrorHandler
	}

	if err = opts.ErrHandler(resp); err != nil {
		return nil, err
	}

	return resp, nil
}
