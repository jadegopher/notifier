package notifier

import (
	"notifier/log"
	"notifier/log/tag"
)

// Notify is semi async func that will be locked if inputChan is full
func (n *Notifier) Notify(msg string) bool {
	if n.isInputChanLocked.Load() {
		log.Warn("Dropping message: inputChan is closed. Graceful shutdown in progress...", tag.Msg, msg)
		return false
	}

	n.inputChan <- msg

	return true
}

// NotifyAndForget drops messages if inputChan is full
func (n *Notifier) NotifyAndForget(msg string) bool {
	if n.isInputChanLocked.Load() {
		log.Warn("Dropping message: inputChan is closed. Graceful shutdown in progress...", tag.Msg, msg)
		return false
	}

	select {
	case n.inputChan <- msg:
		return true
	default:
		log.Warn("Dropping message: inputChan is full", tag.Msg, msg)
		return false
	}
}

// Start is initialization function of notifier. It's necessary to call.
// Start spin up Aggregator and worker pool of SendersCount Senders.
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
	n.isInputChanLocked.Store(true)
	close(n.inputChan)
	n.wg.Wait()
}
