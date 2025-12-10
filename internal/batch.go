package internal

// batch is a non-concurrent safe struct to aggregate messages into batches to send them later to a client via HTTP.
// batch has a limit by byte size. limit can be configured via maxSizeBytes
type batch struct {
	maxSizeBytes int
	sizeBytes    int
	data         []string
}

func newBatch(maxSizeBytes int) *batch {
	return &batch{
		maxSizeBytes: maxSizeBytes,
	}
}

func (b *batch) MaxBatchSizeBytes() int {
	return b.maxSizeBytes
}

func (b *batch) Add(s string) bool {
	addSize := len(s)

	if b.sizeBytes+addSize > b.maxSizeBytes {
		return false
	}

	b.sizeBytes += addSize
	b.data = append(b.data, s)

	return true
}

func (b *batch) Flush() ([]string, int) {
	data := make([]string, len(b.data))
	copy(data, b.data)

	// optimization to reduce slice allocations
	b.data = make([]string, 0, len(b.data))
	sizeBytes := b.sizeBytes
	b.sizeBytes = 0

	return data, sizeBytes
}
