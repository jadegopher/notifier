package notifier

import (
	"notifier/log"
)

type aggregator struct {
	inputChan  <-chan string
	outputChan chan<- []string

	batch []string

	logger log.Logger
}

type batch struct {
	maxSizeBytes int
	sizeBytes    int
	data         []string
}

func (b *batch) Add(s string) bool {
	addSize := len(s)

	if b.sizeBytes+addSize > b.maxSizeBytes {
		return false
	}

	b.data = append(b.data, s)
}
