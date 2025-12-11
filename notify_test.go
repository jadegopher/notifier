package notifier

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestNotifier_End_To_End(t *testing.T) {
	t.Parallel()

	var requestCount int32

	receivedBody := ""

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				// Read body to verify content
				bodyBytes, _ := io.ReadAll(r.Body)
				receivedBody = string(bodyBytes)

				atomic.AddInt32(&requestCount, 1)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()

	n := Default(server.URL)

	n.Start()

	testMsg := "hello_world_integration"
	n.Notify(testMsg)
	n.Stop()

	if diff := cmp.Diff(atomic.LoadInt32(&requestCount), int32(1)); diff != "" {
		t.Errorf("Request count mismatch (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(`{"messages":["hello_world_integration"]}`, receivedBody); diff != "" {
		t.Errorf("Body missmatch (-want +got):\n%s", diff)
	}
}
