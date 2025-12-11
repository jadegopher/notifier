package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"notifier/client"
	"notifier/log"
	"notifier/log/tag"
)

type Sender struct {
	inputChan  <-chan []string
	httpClient client.HTTPClient
	senderFunc func(ctx context.Context, httpClient client.HTTPClient, msg []string) error
}

func encodeBody(_ context.Context, s []string) (io.ReadCloser, error) {
	resp := struct {
		Messages []string `json:"messages"`
	}{Messages: s}

	b, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(b)), nil
}

func NewSender(
	inputChan <-chan []string,
	httpClient client.HTTPClient,
	senderFunc func(ctx context.Context, httpClient client.HTTPClient, msg []string) error,
) *Sender {
	return &Sender{
		inputChan:  inputChan,
		httpClient: httpClient,
		senderFunc: senderFunc,
	}
}

func (s *Sender) Run(id int) {
	log.Debug("sender started", "id", id)

	for msg := range s.inputChan {
		ctx := context.Background()

		if err := s.senderFunc(ctx, s.httpClient, msg); err != nil {
			continue
		}
	}

	log.Debug("sender finished", "id", id)
}

func DefaultSend(ctx context.Context, httpClient client.HTTPClient, msg []string) error {
	body, err := encodeBody(ctx, msg)
	if err != nil {
		log.ErrorContext(ctx, "failed to encode body. dropping msgs", tag.Err, err, tag.Msgs, msg)

		return err
	}

	_, err = httpClient.Do(
		ctx, &http.Request{
			Method: http.MethodPost,
			Body:   body,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "failed to send msgs", tag.Err, err, tag.Msgs, msg)

		return err
	}

	return nil
}
