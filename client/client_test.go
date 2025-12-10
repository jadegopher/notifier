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
		r *http.Response
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
			name:    "bad_request_400_returns_error",
			args:    args{r: newResponse(t, http.StatusBadRequest, `{"error":"invalid fields"}`)},
			wantErr: errs.ErrValidation,
		},
		{
			name:    "not_found_404_returns_error",
			args:    args{r: newResponse(t, http.StatusNotFound, `{"error":"not found"}`)},
			wantErr: errs.ErrNotFound,
		},
		{
			name:    "internal_server_error_500_returns_error",
			args:    args{r: newResponse(t, http.StatusInternalServerError, `{"error":"boom"}`)},
			wantErr: errs.ErrInternal,
		},
		{
			name: "two_response",
			args: args{
				r: newResponse(t, http.StatusInternalServerError, `{"error":"boom"}`),
			},
			wantErr: errs.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				err := DefaultErrorHandler(tt.args.r)
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("DefaultErrorHandler() error = %v, wantErr %v", err, tt.wantErr)
				}
			},
		)
	}
}
