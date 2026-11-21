package amqp

import (
	"context"
	"errors"
	"matt-andrews/postern/publishers/interfaces"
	"sync"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

//----mocks----//

type sentMessage struct {
	queue string
	body  string
}

type fakeChannel struct {
	mu         sync.Mutex
	closed     bool
	declared   []string
	sent       []sentMessage
	declareErr error
	publishErr error
}

func (c *fakeChannel) QueueDeclare(name string, _, _, _, _ bool, _ amqp.Table) (amqp.Queue, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.declareErr != nil {
		return amqp.Queue{}, c.declareErr
	}
	c.declared = append(c.declared, name)
	return amqp.Queue{Name: name}, nil
}

func (c *fakeChannel) PublishWithContext(_ context.Context, _, key string, _, _ bool, msg amqp.Publishing) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.publishErr != nil {
		return c.publishErr
	}
	c.sent = append(c.sent, sentMessage{queue: key, body: string(msg.Body)})
	return nil
}

func (c *fakeChannel) IsClosed() bool { return c.closed }

type fakeConn struct {
	closed bool
	ch     *fakeChannel
}

func (c *fakeConn) Channel() (amqpChannel, error) { return c.ch, nil }
func (c *fakeConn) IsClosed() bool                { return c.closed }
func (c *fakeConn) Close() error                  { c.closed = true; return nil }

// fakeDialer hands out fakeConns and remembers every one it created.
type fakeDialer struct {
	mu    sync.Mutex
	conns []*fakeConn
	uris  []string
	err   error
}

func (d *fakeDialer) dial(uri string) (amqpConn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.err != nil {
		return nil, d.err
	}
	conn := &fakeConn{ch: &fakeChannel{}}
	d.conns = append(d.conns, conn)
	d.uris = append(d.uris, uri)
	return conn, nil
}

type fakeConfig struct{}

func (f fakeConfig) Init(cfg map[string]any, bindings map[string]any) interfaces.PublisherConfig {
	return fakeConfig{}
}

//-----tests-----//

// newTestPublisher returns a publisher wired to a fake dialer.
func newTestPublisher(t *testing.T) (*AmqpPublisher, *fakeDialer) {
	t.Helper()
	d := &fakeDialer{}
	p := NewAmqpPublisher()
	p.dial = d.dial
	return p, d
}

func cfg(uri, topic, message string) AmqpConfig {
	return AmqpConfig{uri: uri, topic: topic, message: message}
}

func TestKey(t *testing.T) {
	if got := NewAmqpPublisher().Key(); got != "amqp" {
		t.Errorf("Key() = %q, want %q", got, "amqp")
	}
}

func TestPublish_SendsMessageToQueue(t *testing.T) {
	p, d := newTestPublisher(t)

	if err := p.Publish(cfg("amqp://a", "orders", "hello")); err != nil {
		t.Fatalf("Publish returned error: %v", err)
	}

	sent := d.conns[0].ch.sent
	if len(sent) != 1 {
		t.Fatalf("got %d messages, want 1", len(sent))
	}
	want := sentMessage{queue: "orders", body: "hello"}
	if sent[0] != want {
		t.Errorf("sent %+v, want %+v", sent[0], want)
	}
}

// Table-driven test: one loop body, many named cases.
func TestPublish_ReusesConnectionPerURI(t *testing.T) {
	tests := []struct {
		name      string
		uris      []string
		wantDials int
	}{
		{"single publish", []string{"amqp://a"}, 1},
		{"same uri twice", []string{"amqp://a", "amqp://a"}, 1},
		{"two brokers", []string{"amqp://a", "amqp://b"}, 2},
		{"interleaved", []string{"amqp://a", "amqp://b", "amqp://a", "amqp://b"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, d := newTestPublisher(t)

			for _, uri := range tt.uris {
				if err := p.Publish(cfg(uri, "q", "m")); err != nil {
					t.Fatalf("Publish(%s): %v", uri, err)
				}
			}

			if got := len(d.conns); got != tt.wantDials {
				t.Errorf("dialed %d times, want %d", got, tt.wantDials)
			}
		})
	}
}

func TestPublish_DeclaresEachQueueOnce(t *testing.T) {
	p, d := newTestPublisher(t)

	for _, topic := range []string{"orders", "orders", "invoices", "orders"} {
		if err := p.Publish(cfg("amqp://a", topic, "m")); err != nil {
			t.Fatalf("Publish: %v", err)
		}
	}

	got := d.conns[0].ch.declared
	if len(got) != 2 || got[0] != "orders" || got[1] != "invoices" {
		t.Errorf("declared %v, want [orders invoices]", got)
	}
}

func TestPublish_RedialsAfterConnectionDrops(t *testing.T) {
	p, d := newTestPublisher(t)

	if err := p.Publish(cfg("amqp://a", "orders", "first")); err != nil {
		t.Fatalf("first Publish: %v", err)
	}

	d.conns[0].closed = true // simulate the broker going away

	if err := p.Publish(cfg("amqp://a", "orders", "second")); err != nil {
		t.Fatalf("second Publish: %v", err)
	}

	if len(d.conns) != 2 {
		t.Fatalf("dialed %d times, want 2", len(d.conns))
	}
	// The queue must be re-declared on the new connection.
	if got := d.conns[1].ch.declared; len(got) != 1 {
		t.Errorf("new connection declared %v, want [orders]", got)
	}
}

func TestPublish_ReopensClosedChannel(t *testing.T) {
	p, d := newTestPublisher(t)

	if err := p.Publish(cfg("amqp://a", "orders", "first")); err != nil {
		t.Fatalf("first Publish: %v", err)
	}

	// Channel dies but the connection survives.
	d.conns[0].ch.closed = true
	d.conns[0].ch = &fakeChannel{}

	if err := p.Publish(cfg("amqp://a", "orders", "second")); err != nil {
		t.Fatalf("second Publish: %v", err)
	}

	if len(d.conns) != 1 {
		t.Errorf("dialed %d times, want 1 (connection should be kept)", len(d.conns))
	}
	if got := len(d.conns[0].ch.sent); got != 1 {
		t.Errorf("new channel got %d messages, want 1", got)
	}
}

func TestPublish_DialErrorIsReturnedAndRetried(t *testing.T) {
	p, d := newTestPublisher(t)
	dialErr := errors.New("connection refused")
	d.err = dialErr

	err := p.Publish(cfg("amqp://a", "orders", "m"))
	if !errors.Is(err, dialErr) {
		t.Fatalf("got error %v, want it to wrap %v", err, dialErr)
	}

	d.err = nil // broker comes back
	if err := p.Publish(cfg("amqp://a", "orders", "m")); err != nil {
		t.Errorf("Publish after recovery: %v", err)
	}
}

func TestPublish_FailedDeclareIsNotCached(t *testing.T) {
	p, d := newTestPublisher(t)

	// Prime the connection so we can reach its channel.
	if err := p.Publish(cfg("amqp://a", "warmup", "m")); err != nil {
		t.Fatalf("warmup: %v", err)
	}
	ch := d.conns[0].ch

	ch.declareErr = errors.New("access refused")
	if err := p.Publish(cfg("amqp://a", "orders", "m")); err == nil {
		t.Fatal("expected an error, got nil")
	}

	ch.declareErr = nil
	if err := p.Publish(cfg("amqp://a", "orders", "m")); err != nil {
		t.Fatalf("Publish after declare recovered: %v", err)
	}
	if got := ch.declared; len(got) != 2 || got[1] != "orders" {
		t.Errorf("declared %v, want [warmup orders]", got)
	}
}

func TestPublish_PublishErrorIsReturned(t *testing.T) {
	p, d := newTestPublisher(t)
	if err := p.Publish(cfg("amqp://a", "orders", "m")); err != nil {
		t.Fatalf("warmup: %v", err)
	}

	pubErr := errors.New("channel blocked")
	d.conns[0].ch.publishErr = pubErr

	if err := p.Publish(cfg("amqp://a", "orders", "m")); !errors.Is(err, pubErr) {
		t.Errorf("got error %v, want it to wrap %v", err, pubErr)
	}
}

func TestPublish_RejectsWrongConfigType(t *testing.T) {
	p, d := newTestPublisher(t)

	if err := p.Publish(fakeConfig{}); err == nil {
		t.Error("expected an error, got nil")
	}
	if len(d.conns) != 0 {
		t.Error("should not dial when the config is invalid")
	}
}

func TestClose_ClosesEveryConnection(t *testing.T) {
	p, d := newTestPublisher(t)
	for _, uri := range []string{"amqp://a", "amqp://b"} {
		if err := p.Publish(cfg(uri, "q", "m")); err != nil {
			t.Fatalf("Publish: %v", err)
		}
	}

	if err := p.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	for i, conn := range d.conns {
		if !conn.closed {
			t.Errorf("connection %d (%s) was not closed", i, d.uris[i])
		}
	}
}

func TestClose_WithNoConnectionsIsSafe(t *testing.T) {
	if err := NewAmqpPublisher().Close(); err != nil {
		t.Errorf("Close on unused publisher: %v", err)
	}
}
