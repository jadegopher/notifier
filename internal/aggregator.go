package internal

import (
	"time"

	"notifier/log"
)

const (
	FlushReasonFull     = "full"
	FlushReasonTimer    = "timer"
	FlushReasonShutdown = "shutdown"

	maxBatchSizeBytesTag = "max_batch_size_b"
)

type Aggregator struct {
	// input channel with messages
	inputChan <-chan string
	// output channel with batched messages
	outputChan chan []string

	// if batch cannot be flushed by overflow condition
	// (number of incoming events too low) then we flush periodically by timer
	flushInterval time.Duration

	batch *batch
}

func NewAggregator(
	inputChan <-chan string,
	outputChanSize int,
	maxBatchSizeBytes int,
	flushInterval time.Duration,
) *Aggregator {
	return &Aggregator{
		inputChan:     inputChan,
		outputChan:    make(chan []string, outputChanSize),
		batch:         newBatch(maxBatchSizeBytes),
		flushInterval: flushInterval,
	}
}

func (a *Aggregator) OutputChan() <-chan []string {
	return a.outputChan
}

func (a *Aggregator) Handle() {
	timer := time.NewTimer(a.flushInterval)
	defer timer.Stop()

	for {
		select {
		case msg, ok := <-a.inputChan:
			if !ok {
				a.flush(FlushReasonShutdown)
				a.finishAggregator()
				return
			}

			if a.batch.Add(msg) {
				continue
			}

			a.flush(FlushReasonFull)
			resetTimer(timer, a.flushInterval)

			if !a.batch.Add(msg) {
				log.Error(
					"failed to add message after flush. msg not sent",
					maxBatchSizeBytesTag, a.batch.maxSizeBytes,
				)

				continue
			}

		case <-timer.C:
			a.flush(FlushReasonTimer)
			resetTimer(timer, a.flushInterval)
		}
	}
}

func resetTimer(timer *time.Timer, flushInterval time.Duration) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
	timer.Reset(flushInterval)
}

func (a *Aggregator) flush(reason string) {
	data, sizeBytes := a.batch.Flush()
	if sizeBytes == 0 {
		return
	}

	log.Debug(
		"batch flushing",
		"reason", reason, "batch_size_b", sizeBytes, maxBatchSizeBytesTag, a.batch.MaxBatchSizeBytes(),
		"flush_period_ms", a.flushInterval.Milliseconds(),
	)

	a.outputChan <- data
}

func (a *Aggregator) finishAggregator() {
	log.Debug("Aggregator: graceful shutdown in progress...")
	close(a.outputChan)
	log.Debug("Aggregator: finished")
}
