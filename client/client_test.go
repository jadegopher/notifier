package client

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"notifier/errs"
)

func newResponse(t *testing.T, status int, body string) *http.Response {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
}

func TestDefaultErrorHandler(t *testing.T) {
	t.Parallel()

	type args struct {
		r   *http.Response
		err error
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name:    "ok_200_returns_nil",
			args:    args{r: newResponse(t, http.StatusOK, ``)},
			wantErr: nil,
		},
		{
			name:    "no_content_204_returns_nil",
			args:    args{r: newResponse(t, http.StatusNoContent, ``)},
			wantErr: nil,
		},
		{
			name:    "redirect_302_returns_nil",
			args:    args{r: newResponse(t, http.StatusFound, ``)},
			wantErr: nil,
		},
		{
			name:    "bad_request_400_returns_validation_error",
			args:    args{r: newResponse(t, http.StatusBadRequest, `{"error":"invalid fields"}`)},
			wantErr: errs.ErrValidation,
		},
		{
			name:    "not_found_404_returns_not_found_error",
			args:    args{r: newResponse(t, http.StatusNotFound, `{"error":"not found"}`)},
			wantErr: errs.ErrNotFound,
		},
		{
			name:    "internal_server_error_500_returns_internal_error",
			args:    args{r: newResponse(t, http.StatusInternalServerError, `{"error":"boom"}`)},
			wantErr: errs.ErrInternal,
		},
		{
			// Tests the "default" switch case
			name:    "forbidden_403_returns_internal_error",
			args:    args{r: newResponse(t, http.StatusForbidden, ``)},
			wantErr: errs.ErrInternal,
		},
		{
			// Tests the "default" switch case
			name:    "teapot_418_returns_internal_error",
			args:    args{r: newResponse(t, http.StatusTeapot, ``)},
			wantErr: errs.ErrInternal,
		},

		{
			// Case where 'Do' failed before sending request (e.g. context timeout)
			name: "preexisting_input_error_returns_it",
			args: args{
				r:   nil,
				err: context.DeadlineExceeded,
			},
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				// Pass both the response and the input error
				err := DefaultErrorHandler(tt.args.r, tt.args.err)

				if !errors.Is(err, tt.wantErr) {
					t.Errorf("DefaultErrorHandler() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}
