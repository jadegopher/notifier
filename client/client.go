package client

import (
	"context"
	"net/http"

	"notifier/errs"
	"notifier/log"
	"notifier/log/tag"
)

// HTTPClient decouples dependency on specific HTTP requesting library.
type HTTPClient interface {
	Do(ctx context.Context, req *http.Request, opts Options) (*http.Response, error)
}

type ErrorHandler func(r *http.Response) error

type Options struct {
	// If not provided, DefaultErrorHandler will be used.
	ErrHandler ErrorHandler
}

func DefaultErrorHandler(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode <= 399 {
		return nil
	}

	switch r.StatusCode {
	case http.StatusNotFound:
		return errs.Wrap(errs.ErrNotFound, r.Request.URL.Path)
	case http.StatusBadRequest:
		log.Error(r.Request.Context(), "bad request", tag.HTTPCode, r.StatusCode)
		return errs.Wrap(errs.ErrValidation, r.Request.URL.Path)
	default:
		log.Error(r.Request.Context(), "unexpected status code", tag.HTTPCode, r.StatusCode)
		return errs.Wrap(errs.ErrInternal, r.Request.URL.Path)
	}
}
