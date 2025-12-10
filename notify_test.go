package notifier

import (
	"sync"
	"testing"

	"notifier/internal"
)

func TestNotifier_Notify(t *testing.T) {
	t.Parallel()

	type fields struct {
		inputChan    chan string
		aggregator   *internal.Aggregator
		sendersCount int
		sender       *internal.Sender
		wg           *sync.WaitGroup
	}

	type args struct {
		msg string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				n := &Notifier{
					inputChan:    tt.fields.inputChan,
					aggregator:   tt.fields.aggregator,
					sendersCount: tt.fields.sendersCount,
					sender:       tt.fields.sender,
					wg:           tt.fields.wg,
				}
				n.Notify(tt.args.msg)
			},
		)
	}
}
