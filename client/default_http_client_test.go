package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"
)

func TestDefaultHTTPClient_Do(t *testing.T) {
	t.Parallel()

	type fields struct {
		c            *resty.Client
		limiter      *rate.Limiter
		errorHandler ErrorHandler
	}

	type args struct {
		ctx    context.Context
		method string
	}

	tests := []struct {
		name          string
		serverHandler func(w http.ResponseWriter, r *http.Request)
		fields        fields
		args          args
		wantStatus    int
		wantErr       bool
	}{
		{
			name: "success_happy_path",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			fields: fields{
				c:       resty.New(),
				limiter: rate.NewLimiter(rate.Inf, 0),
			},
			args: args{
				ctx:    context.Background(),
				method: http.MethodGet,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "retry_success_after_failure",
			// Server logic: Fail 2 times with 500, then succeed with 200
			serverHandler: func() func(w http.ResponseWriter, r *http.Request) {
				var calls int32
				return func(w http.ResponseWriter, r *http.Request) {
					count := atomic.AddInt32(&calls, 1)
					if count <= 2 {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
				}
			}(),
			fields: fields{
				c: resty.New().
					SetRetryCount(3).
					SetRetryWaitTime(1 * time.Millisecond).
					AddRetryCondition(DefaultRetryCondition),
				limiter: rate.NewLimiter(rate.Inf, 0),
			},
			args: args{
				ctx:    context.Background(),
				method: http.MethodGet,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "retry_exhausted_fail",
			// Server logic: Always fail
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			fields: fields{
				c: resty.New().
					SetRetryCount(2).
					SetRetryWaitTime(1 * time.Millisecond).
					// Even with retry condition, this will eventually fail
					AddRetryCondition(DefaultRetryCondition),
				limiter: rate.NewLimiter(rate.Inf, 0),
			},
			args: args{
				ctx:    context.Background(),
				method: http.MethodGet,
			},
			wantStatus: 0,
			wantErr:    true,
		},
		{
			name: "rate_limit_exceeded_error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			fields: fields{
				c:       resty.New(),
				limiter: rate.NewLimiter(rate.Every(1*time.Hour), 1),
			},
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
					defer cancel()
					return ctx
				}(),
				method: http.MethodGet,
			},
			wantStatus: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				srv := httptest.NewServer(http.HandlerFunc(tt.serverHandler))
				defer srv.Close()

				if tt.name == "rate_limit_exceeded_error" {
					// Consume the single burst token so the next one blocks
					tt.fields.limiter.Allow()
				}

				r := NewDefaultHTTPClient(tt.fields.c, tt.fields.limiter, tt.fields.errorHandler)

				req, err := http.NewRequest(tt.args.method, srv.URL, nil)
				if err != nil {
					t.Fatalf("Failed to create request: %v", err)
				}

				got, err := r.Do(tt.args.ctx, req)

				if (err != nil) != tt.wantErr {
					t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if !tt.wantErr {
					if got == nil {
						t.Fatal("Do() returned nil response but expected success")
					}
					if got.StatusCode != tt.wantStatus {
						t.Errorf("Do() status = %v, want %v", got.StatusCode, tt.wantStatus)
					}
				}
			},
		)
	}
}
