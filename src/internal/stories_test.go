//go:build integration

package internal

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"matt-andrews/postern/internal/harness"
	"matt-andrews/postern/publishers"
	"matt-andrews/postern/runtime"
)

// runs postern in-process with the given YAML config and returns its base URL, e.g. "http://127.0.0.1:54321".
func startPostern(t *testing.T, yamlConfig string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "postern.yaml")
	if err := os.WriteFile(path, []byte(yamlConfig), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	publishers.Init()
	srv := httptest.NewServer(runtime.NewRouter(runtime.Discover(path)))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestStory_HttpBecomesAmqpMessage(t *testing.T) {
	broker := harness.StartRabbit(t)
	queue := harness.UniqueName("orders")

	postern := startPostern(t, fmt.Sprintf(`
matches:
  - name: order-created
    route:
      path: "/hooks/:topic"
      method: POST
    out:
      - type: amqp
        uri: %q
        topic: "{{ route.topic }}"
        message: '{"order": "{{ json.id }}"}'
`, broker.URI))

	resp, err := http.Post(postern+"/hooks/"+queue, "application/json", strings.NewReader(`{"id":"abc"}`))
	if err != nil {
		t.Fatalf("post webhook: %v", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	msg := broker.ExpectMessage(t, queue, 5*time.Second)
	if want := `{"order": "abc"}`; msg.Body != want {
		t.Errorf("message body = %s, want %s", msg.Body, want)
	}
}

func TestStory_HttpBecomesKafkaMessage(t *testing.T) {
	broker := harness.StartKafka(t)
	topic := broker.CreateTopic(t, "orders")

	postern := startPostern(t, fmt.Sprintf(`
matches:
  - name: order-created
    route:
      path: "/hooks/:topic"
      method: POST
    out:
      - type: kafka
        brokers: [%q]
        topic: "{{ route.topic }}"
        key: "{{ json.customer_id }}"
        message: '{"order": {{ json.id | json }}}'
        headers:
          source: "postern"
`, broker.Brokers[0]))

	resp, err := http.Post(postern+"/hooks/"+topic, "application/json", strings.NewReader(`{"id":"abc","customer_id":"c-42"}`))
	if err != nil {
		t.Fatalf("post webhook: %v", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	msg := broker.ExpectMessage(t, topic, 10*time.Second)
	if want := `{"order": "abc"}`; msg.Body != want {
		t.Errorf("message body = %s, want %s", msg.Body, want)
	}
	if want := "c-42"; msg.Key != want {
		t.Errorf("message key = %q, want %q", msg.Key, want)
	}
	if want := "postern"; msg.Headers["source"] != want {
		t.Errorf("header source = %q, want %q", msg.Headers["source"], want)
	}
}
