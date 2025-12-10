package errs

import (
	"fmt"
)

var (
	ErrValidation = fmt.Errorf("validation error")
	ErrInternal   = fmt.Errorf("internal error")
	ErrNotFound   = fmt.Errorf("not found")
)

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	if msg == "" {
		return err
	}

	return fmt.Errorf("%s: %w", msg, err)
}
