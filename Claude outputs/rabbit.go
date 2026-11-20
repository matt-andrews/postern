// Package harness holds helpers for tests that need a real broker.
// It lives under internal/ so only postern's own packages can import it.
package harness

import (
	"context"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

// Rabbit is a RabbitMQ broker running in a container for the duration of a test.
type Rabbit struct {
	// URI is what goes in the `uri:` field of a postern config.
	URI string

	conn *amqp.Connection
}

// Message is what a test gets back when it expects a message.
type Message struct {
	Body        string
	ContentType string
}

// StartRabbit starts a RabbitMQ container and stops it when the test ends.
func StartRabbit(t testing.TB) *Rabbit {
	t.Helper()
	ctx := context.Background()

	ctr, err := rabbitmq.Run(ctx, "rabbitmq:4-management")
	testcontainers.CleanupContainer(t, ctr)
	if err != nil {
		t.Fatalf("start rabbitmq: %v", err)
	}

	uri, err := ctr.AmqpURL(ctx)
	if err != nil {
		t.Fatalf("rabbitmq url: %v", err)
	}

	conn, err := amqp.Dial(uri)
	if err != nil {
		t.Fatalf("dial rabbitmq: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	return &Rabbit{URI: uri, conn: conn}
}

// ExpectMessage waits for one message on queue and returns it.
// It fails the test if nothing arrives before the timeout.
func (r *Rabbit) ExpectMessage(t testing.TB, queue string, timeout time.Duration) Message {
	t.Helper()

	// A channel of our own per call: closing it hands back anything we
	// were sent but did not ack, so the next ExpectMessage still sees it.
	ch, err := r.conn.Channel()
	if err != nil {
		t.Fatalf("open channel: %v", err)
	}
	defer ch.Close()

	deliveries, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		// Usually "NOT_FOUND": postern never declared the queue,
		// which means it never got as far as publishing.
		t.Fatalf("consume from %q: %v", queue, err)
	}

	select {
	case d := <-deliveries:
		if err := d.Ack(false); err != nil {
			t.Fatalf("ack message from %q: %v", queue, err)
		}
		return Message{Body: string(d.Body), ContentType: d.ContentType}
	case <-time.After(timeout):
		t.Fatalf("no message on %q after %s", queue, timeout)
		return Message{}
	}
}
