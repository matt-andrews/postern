//go:build integration

package main

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
)

// runs postern in-process with the given YAML config and returns its base URL, e.g. "http://127.0.0.1:54321".
func startPostern(t *testing.T, yamlConfig string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "postern.yaml")
	if err := os.WriteFile(path, []byte(yamlConfig), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	publishers.Init()
	srv := httptest.NewServer(newRouter(Discover(path)))
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
