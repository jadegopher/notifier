package client

import (
	"context"
	"net/http"

	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"

	"notifier/errs"
)

type DefaultHTTPClient struct {
	c            *resty.Client
	limiter      *rate.Limiter
	errorHandler ErrorHandler
}

func NewDefaultHTTPClient(c *resty.Client, limiter *rate.Limiter, errorHandler ErrorHandler) HTTPClient {
	d := &DefaultHTTPClient{c: c, limiter: limiter, errorHandler: errorHandler}

	if d.errorHandler == nil {
		d.errorHandler = DefaultErrorHandler
	}

	return d
}

func (r *DefaultHTTPClient) Do(
	ctx context.Context,
	req *http.Request,
) (*http.Response, error) {
	err := r.limiter.Wait(ctx)
	if err = r.errorHandler(nil, errs.Wrap(err, "rate limiter")); err != nil {
		return nil, err
	}

	// Create a Resty request and copy fields from stdReq
	restyReq := r.c.R().
		SetContext(req.Context()).
		SetHeaderMultiValues(req.Header).SetBody(req.Body)

	resp, err := restyReq.Execute(req.Method, req.URL.String())
	if err = r.errorHandler(getRawResponse(resp), err); err != nil {
		return nil, err
	}

	return getRawResponse(resp), nil
}

func getRawResponse(resp *resty.Response) *http.Response {
	if resp == nil {
		return nil
	}

	return resp.RawResponse
}
