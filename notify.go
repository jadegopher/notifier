package notifier

import (
	"notifier/log"
)

func Notify(msg string) {
	DefaultNotifier.Notify(msg)
}

func (n *Notifier) Notify(msg string) {
	// closed channel + buffr overflow
	n.inputChan <- msg
}

func (n *Notifier) Start() {
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()
		n.aggregator.Handle()
	}()

	for i := 0; i < n.sendersCount; i++ {
		n.wg.Add(1)

		go func(id int) {
			defer n.wg.Done()

			n.sender.Run(id)
		}(i)
	}
}

func (n *Notifier) Stop() {
	log.Debug("Notifier: Graceful shutdown in progress...")
	close(n.inputChan)
	n.wg.Wait()
}
