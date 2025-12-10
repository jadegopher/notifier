package notifier

import (
	"context"
)

func Notify(msg string) {
	DefaultNotifier.Notify(msg)
}

func (n *Notifier) Notify(msg string) {
	n.inputChan <- msg
}

func (n *Notifier) Init(ctx context.Context) {
	//n.errGroup = &errgroup.Group{}
	n.errGroup.Go(n.handle)

	//
	n.errGroup.Go(
		func() error {
			<-ctx.Done()

			n.logger.Debug("Notifier: context done received")
		},
	)
}

func (n *Notifier) handle() error {
	select {
	case msg, ok := <-n.inputChan:
		if !ok {

		}
	}

	return nil
}

func (n *Notifier) flush() {

}

func (n *Notifier) finishAggregator() {
	n.logger.Debug("Notifier.Aggregator: graceful shutdown in progress...")
	close(n.outputChan)
	n.logger.Debug("Notifier.Aggregator: finished")
}
