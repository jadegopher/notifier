package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"notifier/client"
	"notifier/log"
	"notifier/log/tag"
)

type Sender struct {
	inputChan        <-chan []string
	url              *url.URL
	httpClient       client.HTTPClient
	httpErrorHandler client.ErrorHandler
	bodyEncoder      func(context.Context, []string) (io.ReadCloser, error)
}

func DefaultBodyEncoder() func(context.Context, []string) (io.ReadCloser, error) {
	return func(ctx context.Context, s []string) (io.ReadCloser, error) {
		resp := struct {
			Messages []string `json:"messages"`
		}{Messages: s}

		b, err := json.Marshal(resp)
		if err != nil {
			return nil, err
		}

		return io.NopCloser(bytes.NewReader(b)), nil
	}
}

func NewSender(
	inputChan <-chan []string,
	httpClient client.HTTPClient,
	bodyEncoder func(context.Context, []string) (io.ReadCloser, error),
) *Sender {
	return &Sender{
		inputChan:   inputChan,
		httpClient:  httpClient,
		bodyEncoder: bodyEncoder,
	}
}

func (s *Sender) Run(id int) {
	log.Debug("sender started", "id", id)

	for msg := range s.inputChan {
		ctx := context.Background()

		body, err := s.bodyEncoder(ctx, msg)
		if err != nil {
			log.ErrorContext(ctx, "failed to encode body. dropping msgs", tag.Err, err, tag.Msgs, msg)

			continue
		}

		_, err = s.httpClient.Do(
			ctx, &http.Request{
				Method: http.MethodPost,
				URL:    s.url,
				Body:   body,
			},
		)
		if err != nil {
			log.ErrorContext(ctx, "failed to send msgs", tag.Err, err, tag.Msgs, msg)

			continue
		}
	}

	log.Debug("sender finished", "id", id)
}
