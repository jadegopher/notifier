package errs

import (
	"errors"
	"testing"
)

func TestWrap(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
		msg string
	}

	tests := []struct {
		name     string
		args     args
		wantErr  bool
		wantText string
	}{
		{
			name: "nil_error",
			args: args{
				err: nil,
				msg: "",
			},
			wantErr:  false,
			wantText: "",
		},
		{
			name: "nil_error_w_text",
			args: args{
				err: nil,
				msg: "text",
			},
			wantErr:  false,
			wantText: "",
		},
		{
			name: "error_with_no_text",
			args: args{
				err: ErrValidation,
				msg: "",
			},
			wantErr:  true,
			wantText: "validation error",
		},
		{
			name: "error_with_text",
			args: args{
				err: ErrValidation,
				msg: "id is required",
			},
			wantErr:  true,
			wantText: "id is required: validation error",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				err := Wrap(tt.args.err, tt.args.msg)
				if (err != nil) != tt.wantErr {
					t.Errorf("Wrap() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !tt.wantErr {
					return
				}

				if !errors.Is(err, tt.args.err) {
					t.Errorf("errors.Is is false %v, %v", err, tt.args.err)
				}

				if err.Error() != tt.wantText {
					t.Errorf("error text not equal %s %s", err.Error(), tt.wantText)
				}
			},
		)
	}
}
