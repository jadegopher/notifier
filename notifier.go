package notifier

import (
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"

	"notifier/client"
	"notifier/internal"
)

const (
	// DefaultInputChanSize sets input channel size
	DefaultInputChanSize = 5000

	// DefaultBatchSizeBytes aggregator will batch messages from input channel to batches of size N
	DefaultBatchSizeBytes = 1 * 1024 * 1024 // 1 MB
	// DefaultFlushInterval aggregator will drop every N seconds messages if batch not filled
	DefaultFlushInterval = 1 * time.Second
	// DefaultOutputChanSize sets the size of a channel that sends batched messages to senders
	DefaultOutputChanSize = DefaultSendersCount * 10

	// DefaultSendersCount sets the number of workers that will send batched messages to specified URL
	DefaultSendersCount = 10
	// DefaultHTTPTimeout default timeout for sender
	DefaultHTTPTimeout = 10 * time.Second
	// DefaultRetryCount default count of attempts to retry sending msgs
	DefaultRetryCount = 3
	// DefaultRetryDelay sets the default wait time for sleep before retrying request
	DefaultRetryDelay = 100 * time.Millisecond
	// DefaultRetryMaxDelay method sets the max wait time for sleep before retrying request
	DefaultRetryMaxDelay = 300 * time.Millisecond

	// DefaultRPS sets limit of RPS for senders
	DefaultRPS = 1000
)

var DefaultNotifier = Default()

func Default() *Notifier {
	c := resty.New()
	c.SetTimeout(DefaultHTTPTimeout)

	// Backoff retry mechanism
	c.SetRetryWaitTime(DefaultRetryDelay)
	c.SetRetryMaxWaitTime(DefaultRetryMaxDelay)
	c.SetRetryCount(DefaultRetryCount)

	return NewNotifier(
		client.NewDefaultHTTPClient(
			c,
			rate.NewLimiter(rate.Limit(DefaultRPS), DefaultRPS),
			client.DefaultErrorHandler,
		),
		DefaultInputChanSize,
		DefaultOutputChanSize,
		DefaultBatchSizeBytes,
		DefaultSendersCount,
		DefaultFlushInterval,
	)
}

type Notifier struct {
	inputChan chan string

	aggregator *internal.Aggregator

	sendersCount int
	sender       *internal.Sender

	wg *sync.WaitGroup
}

func NewNotifier(
	httpClient client.HTTPClient,
	inputChanSize int,
	outputChanSize int,
	batchSize int,
	sendersCount int,
	flushInterval time.Duration,
) *Notifier {
	n := &Notifier{
		inputChan:    make(chan string, inputChanSize),
		sendersCount: sendersCount,
		wg:           &sync.WaitGroup{},
	}

	n.aggregator = internal.NewAggregator(n.inputChan, outputChanSize, batchSize, flushInterval)

	n.sender = internal.NewSender(n.aggregator.OutputChan(), httpClient, internal.DefaultBodyEncoder())

	return n
}
