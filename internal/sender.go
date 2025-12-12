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

type SenderFunc func(ctx context.Context, senderID int, httpClient client.HTTPClient, msg []string) error

type Sender struct {
	inputChan  <-chan []string
	httpClient client.HTTPClient
	senderFunc SenderFunc
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
	senderFunc SenderFunc,
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

		if err := s.senderFunc(ctx, id, s.httpClient, msg); err != nil {
			continue
		}
	}

	log.Debug("sender finished", "id", id)
}

func DefaultSend(ctx context.Context, id int, httpClient client.HTTPClient, msg []string) error {
	body, err := encodeBody(ctx, msg)
	if err != nil {
		log.ErrorContext(ctx, "failed to encode body. dropping msgs", tag.ID, id, tag.Err, err, tag.Msgs, len(msg))

		return err
	}

	_, err = httpClient.Do(
		ctx, &http.Request{
			Method: http.MethodPost,
			Body:   body,
		},
	)
	if err != nil {
		log.ErrorContext(ctx, "sender: failed to send msgs", tag.ID, id, tag.Err, err, tag.Msgs, len(msg))

		return err
	}

	log.DebugContext(ctx, "sender: messages sent", tag.ID, id, tag.Msgs, len(msg))

	return nil
}
