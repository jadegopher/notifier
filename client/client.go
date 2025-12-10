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
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

// ErrorHandler Allows the caller to handle notification failures in case any requests fail.
// Check r for nil!
type ErrorHandler func(r *http.Response, err error) error

func DefaultErrorHandler(r *http.Response, err error) error {
	if err != nil {
		log.Error("failed perform request", err)

		return err
	}

	if r == nil {
		return nil
	}

	if r.StatusCode >= 200 && r.StatusCode <= 399 {
		return nil
	}

	switch r.StatusCode {
	case http.StatusNotFound:
		return errs.Wrap(errs.ErrNotFound, r.Request.URL.Path)
	case http.StatusBadRequest:
		log.ErrorContext(r.Request.Context(), "bad request", tag.HTTPCode, r.StatusCode)
		return errs.Wrap(errs.ErrValidation, r.Request.URL.Path)
	default:
		log.ErrorContext(r.Request.Context(), "unexpected status code", tag.HTTPCode, r.StatusCode)
		return errs.Wrap(errs.ErrInternal, r.Request.URL.Path)
	}
}
