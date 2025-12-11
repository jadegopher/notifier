package client

import (
	"context"
	"net/http"

	"github.com/go-resty/resty/v2"
)

type DefaultHTTPClient struct {
	c            *resty.Client
	errorHandler ErrorHandler
}

func DefaultRetryCondition(r *resty.Response, _ error) bool {
	return r.StatusCode() == http.StatusInternalServerError
}

func NewDefaultHTTPClient(c *resty.Client, errorHandler ErrorHandler) HTTPClient {
	d := &DefaultHTTPClient{c: c, errorHandler: errorHandler}

	if d.errorHandler == nil {
		d.errorHandler = DefaultErrorHandler
	}

	return d
}

func (r *DefaultHTTPClient) Do(
	ctx context.Context,
	req *http.Request,
) (*http.Response, error) {
	restyReq := r.c.R().SetContext(ctx).SetBody(req.Body)

	resp, err := restyReq.Execute(req.Method, "")
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
