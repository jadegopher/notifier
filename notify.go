package notifier

import (
	"notifier/log"
)

func (n *Notifier) Notify(msg string) {
	// closed channel + buffer overflow
	n.inputChan <- msg
}

// Start is initialization function of notifier. It's necessary to call.
// Start spin up Aggregator and worker pool of sendersCount Senders.
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

// Stop initiates a graceful shutdown mechanism. It's required to call to finish notifier gracefully.
func (n *Notifier) Stop() {
	log.Debug("Notifier: Graceful shutdown in progress...")
	close(n.inputChan)
	n.wg.Wait()
}
