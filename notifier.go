package notifier

import (
	"sync"
	"sync/atomic"
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

type Options struct {
	InputChanSize  int
	OutputChanSize int
	BatchSize      int
	SendersCount   int
	FlushInterval  time.Duration
}

// Default sets up Notifier with optimal configuration.
// You can additionally tweak some settings and pass them as Options arg.
// However, you can use NewNotifier to tweak almost everything.
func Default(url string, opt ...Options) *Notifier {
	options := parseOptional(opt)

	c := resty.New()
	c.SetTimeout(DefaultHTTPTimeout)
	c.SetBaseURL(url)

	c.SetRateLimiter(rate.NewLimiter(rate.Limit(DefaultRPS), DefaultRPS))

	// Backoff retry mechanism
	c.SetRetryWaitTime(DefaultRetryDelay)
	c.SetRetryMaxWaitTime(DefaultRetryMaxDelay)
	c.SetRetryCount(DefaultRetryCount)
	c.AddRetryCondition(client.DefaultRetryCondition)

	return NewNotifier(
		client.NewDefaultHTTPClient(
			c,
			client.DefaultErrorHandler,
		),
		options.InputChanSize,
		options.OutputChanSize,
		options.BatchSize,
		options.SendersCount,
		options.FlushInterval,
		internal.DefaultSend,
	)
}

type Notifier struct {
	inputChan chan string

	aggregator *internal.Aggregator

	sendersCount int
	sender       *internal.Sender

	isInputChanLocked atomic.Bool

	wg *sync.WaitGroup
}

func NewNotifier(
	httpClient client.HTTPClient,
	inputChanSize int,
	outputChanSize int,
	batchSize int,
	sendersCount int,
	flushInterval time.Duration,
	senderFunc internal.SenderFunc,
) *Notifier {
	n := &Notifier{
		inputChan:         make(chan string, inputChanSize),
		isInputChanLocked: atomic.Bool{},
		sendersCount:      sendersCount,
		wg:                &sync.WaitGroup{},
	}

	n.aggregator = internal.NewAggregator(n.inputChan, outputChanSize, batchSize, flushInterval)

	n.sender = internal.NewSender(n.aggregator.OutputChan(), httpClient, senderFunc)

	return n
}

func parseOptional(opt []Options) Options {
	if len(opt) == 0 {
		return Options{
			InputChanSize:  DefaultInputChanSize,
			OutputChanSize: DefaultOutputChanSize,
			BatchSize:      DefaultBatchSizeBytes,
			SendersCount:   DefaultSendersCount,
			FlushInterval:  DefaultFlushInterval,
		}
	}

	result := opt[0]

	if result.InputChanSize == 0 {
		result.InputChanSize = DefaultInputChanSize
	}
	if result.OutputChanSize == 0 {
		result.OutputChanSize = DefaultOutputChanSize
	}
	if result.BatchSize == 0 {
		result.BatchSize = DefaultBatchSizeBytes
	}
	if result.SendersCount == 0 {
		result.SendersCount = DefaultSendersCount
	}
	if result.FlushInterval == 0 {
		result.FlushInterval = DefaultFlushInterval
	}

	return result
}
