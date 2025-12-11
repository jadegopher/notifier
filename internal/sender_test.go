package internal

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestDefaultBodyEncoder(t *testing.T) {
	t.Parallel()

	type args struct {
		in0 context.Context
		s   []string
	}

	tests := []struct {
		name     string
		args     args
		wantBody string
		wantErr  bool
	}{
		{
			name: "nil_slice_produces_null",
			args: args{
				in0: context.Background(),
				s:   nil,
			},
			wantBody: `{"messages":null}`,
			wantErr:  false,
		},
		{
			name: "empty_slice_produces_empty_json_array",
			args: args{
				in0: context.Background(),
				s:   []string{},
			},
			wantBody: `{"messages":[]}`,
			wantErr:  false,
		},
		{
			name: "single_string",
			args: args{
				in0: context.Background(),
				s:   []string{"hello"},
			},
			wantBody: `{"messages":["hello"]}`,
			wantErr:  false,
		},
		{
			name: "multiple_strings",
			args: args{
				in0: context.Background(),
				s:   []string{"foo", "bar", "baz"},
			},
			wantBody: `{"messages":["foo","bar","baz"]}`,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				got, err := encodeBody(tt.args.in0, tt.args.s)
				if (err != nil) != tt.wantErr {
					t.Errorf("encodeBody() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				// Helper to read the body content for comparison
				readBody := func(r io.Reader) string {
					if r == nil {
						return ""
					}
					b, _ := io.ReadAll(r)
					return strings.TrimSpace(string(b))
				}

				gotBody := readBody(got)
				if gotBody != tt.wantBody {
					t.Errorf("encodeBody() gotBody = %v, want %v", gotBody, tt.wantBody)
				}
			},
		)
	}
}

func BenchmarkDefaultBodyEncoder(b *testing.B) {
	ctx := context.Background()
	input := []string{
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
		"message 1", "message 2", "message 3", "message 4", "message 5",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// We discard the result to avoid compiler optimizations removing the function call,
		// but we don't need to read the body as we are benchmarking the Encoder logic itself.
		_, _ = encodeBody(ctx, input)
	}
}
