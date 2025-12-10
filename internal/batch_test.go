package internal

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_batch_Add(t *testing.T) {
	t.Parallel()

	type args struct {
		s string
	}

	tests := []struct {
		name          string
		fields        batch
		args          args
		want          bool
		wantSizeBytes int      // Expected internal size after operation
		wantData      []string // Expected internal data after operation
	}{
		{
			name: "successfully_add_string_to_empty_batch",
			fields: batch{
				maxSizeBytes: 10,
				sizeBytes:    0,
				data:         []string{},
			},
			args: args{
				s: "hello", // len 5
			},
			want:          true,
			wantSizeBytes: 5,
			wantData:      []string{"hello"},
		},
		{
			name: "successfully_add_string_to_partially_full_batch",
			fields: batch{
				maxSizeBytes: 10,
				sizeBytes:    4,
				data:         []string{"data"},
			},
			args: args{
				s: "world", // len 5, total 9 <= 10
			},
			want:          true,
			wantSizeBytes: 9,
			wantData:      []string{"data", "world"},
		},
		{
			name: "successfully_add_string_that_fits_exactly_(boundary)",
			fields: batch{
				maxSizeBytes: 5,
				sizeBytes:    0,
				data:         []string{},
			},
			args: args{
				s: "abcde", // len 5, total 5 == 5
			},
			want:          true,
			wantSizeBytes: 5,
			wantData:      []string{"abcde"},
		},
		{
			name: "fail_to_add_string_that_exceeds_max_size_(empty_batch)",
			fields: batch{
				maxSizeBytes: 5,
				sizeBytes:    0,
				data:         []string{},
			},
			args: args{
				s: "abcdef", // len 6, total 6 > 5
			},
			want:          false,
			wantSizeBytes: 0,          // Should remain unchanged
			wantData:      []string{}, // Should remain unchanged
		},
		{
			name: "fail_to_add_string_that_exceeds_max_size_(partially_full)",
			fields: batch{
				maxSizeBytes: 10,
				sizeBytes:    9,
				data:         []string{"almost_full"},
			},
			args: args{
				s: "no", // len 2, total 11 > 10
			},
			want:          false,
			wantSizeBytes: 9,                       // Should remain unchanged
			wantData:      []string{"almost_full"}, // Should remain unchanged
		},
		{
			name: "successfully_add_empty_string",
			fields: batch{
				maxSizeBytes: 10,
				sizeBytes:    5,
				data:         []string{"test"},
			},
			args: args{
				s: "", // len 0
			},
			want:          true,
			wantSizeBytes: 5,
			wantData:      []string{"test", ""},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				b := &batch{
					maxSizeBytes: tt.fields.maxSizeBytes,
					sizeBytes:    tt.fields.sizeBytes,
					data:         tt.fields.data,
				}

				if got := b.Add(tt.args.s); got != tt.want {
					t.Errorf("Add() = %v, want %v", got, tt.want)
				}

				if b.sizeBytes != tt.wantSizeBytes {
					t.Errorf("batch.sizeBytes = %v, want %v", b.sizeBytes, tt.wantSizeBytes)
				}

				if diff := cmp.Diff(b.data, tt.wantData); diff != "" {
					t.Errorf("diff %s", diff)
				}
			},
		)
	}
}

func Test_batch_Flush(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fields   batch
		wantData []string // The data returned by Flush
		wantSize int      // The size returned by Flush
	}{
		{
			name: "flush_populated_batch",
			fields: batch{
				maxSizeBytes: 100,
				sizeBytes:    10,
				data:         []string{"hello", "world"},
			},
			wantData: []string{"hello", "world"},
			wantSize: 10,
		},
		{
			name: "flush_empty_batch",
			fields: batch{
				maxSizeBytes: 100,
				sizeBytes:    0,
				data:         []string{},
			},
			wantData: []string{},
			wantSize: 0,
		},
		{
			name: "flush_nil_data_batch",
			fields: batch{
				maxSizeBytes: 100,
				sizeBytes:    0,
				data:         nil,
			},
			wantData: []string{}, // copy() on nil creates empty slice
			wantSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				b := &batch{
					maxSizeBytes: tt.fields.maxSizeBytes,
					sizeBytes:    tt.fields.sizeBytes,
					data:         tt.fields.data,
				}

				gotData, gotSize := b.Flush()

				if !reflect.DeepEqual(gotData, tt.wantData) {
					t.Errorf("Flush() data = %v, want %v", gotData, tt.wantData)
				}
				if gotSize != tt.wantSize {
					t.Errorf("Flush() size = %v, want %v", gotSize, tt.wantSize)
				}

				if b.sizeBytes != 0 {
					t.Errorf("After Flush(), b.sizeBytes = %v, want 0", b.sizeBytes)
				}
				if len(b.data) != 0 {
					t.Errorf("After Flush(), len(b.data) = %v, want 0", len(b.data))
				}

				expectedCap := len(tt.fields.data)
				if cap(b.data) != expectedCap {
					t.Errorf("After Flush(), cap(b.data) = %v, want %v", cap(b.data), expectedCap)
				}
			},
		)
	}
}
