package notifier

import (
	"log/slog"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"

	"notifier/client"
	"notifier/log"
)

const (
	DefaultInputChanSize  = 100
	DefaultHTTPTimeout    = 30 * time.Second
	DefaultBatchSizeBytes = 3 * 1024 // 3 KB
)

var DefaultNotifier = Default()

func Default() *Notifier {
	return NewNotifier(
		client.NewDefaultHTTPClient(
			&http.Client{
				Timeout: DefaultHTTPTimeout,
			},
		),
		client.DefaultErrorHandler,
		DefaultInputChanSize,
		DefaultBatchSizeBytes,
		slog.Default(),
	)
}

type Notifier struct {
	httpClient       client.HTTPClient
	httpErrorHandler client.ErrorHandler

	inputChan chan string

	batchSizeBytes int
	outputChan     chan []string

	logger   log.Logger
	errGroup *errgroup.Group
}

func NewNotifier(
	httpClient client.HTTPClient,
	httpErrorHandler client.ErrorHandler,
	inputChanSize int,
	batchSize int,
	logger log.Logger,
) *Notifier {
	return &Notifier{
		httpClient:       httpClient,
		httpErrorHandler: httpErrorHandler,
		inputChan:        make(chan string, inputChanSize),
		batchSizeBytes:   batchSize,
		logger:           logger,
		errGroup:         &errgroup.Group{},
	}
}
