package publishers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AmqpPublisher struct {
	mu      sync.Mutex
	brokers map[string]*broker
}

func NewAmqpPublisher() *AmqpPublisher {
	return &AmqpPublisher{brokers: make(map[string]*broker)}
}

func (p *AmqpPublisher) Key() string {
	return "amqp"
}

func (p *AmqpPublisher) Config(cfg map[string]string, bindings map[string]any) PublisherConfig {
	return AmpqConfig{}.Init(cfg, bindings)
}

func (p *AmqpPublisher) Publish(config PublisherConfig) error {
	cfg, ok := config.(AmpqConfig)
	if !ok {
		panic("Config type was expected to be AmpqConfig")
	}

	b := p.getBroker(cfg.uri)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return b.publish(ctx, cfg.topic, []byte(cfg.message))
}

func (p *AmqpPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var errs []error
	for _, b := range p.brokers {
		errs = append(errs, b.close())
	}
	return errors.Join(errs...)
}
func (p *AmqpPublisher) getBroker(uri string) *broker {
	p.mu.Lock()
	defer p.mu.Unlock()

	b, ok := p.brokers[uri]
	if !ok {
		b = &broker{uri: uri}
		p.brokers[uri] = b
	}
	return b
}

type broker struct {
	uri string

	mu       sync.Mutex
	conn     *amqp.Connection
	ch       *amqp.Channel
	declared map[string]bool
}

func (b *broker) publish(ctx context.Context, queue string, body []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if err := b.ensureChannel(); err != nil {
		return err
	}

	if !b.declared[queue] {
		if _, err := b.ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare queue %q: %w", queue, err)
		}
		b.declared[queue] = true
	}

	err := b.ch.PublishWithContext(ctx, "", queue, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         body,
	})
	if err != nil {
		return fmt.Errorf("publish to %q: %w", queue, err)
	}
	return nil
}

// ensureChannel dials or redials as needed. Caller must hold b.mu.
func (b *broker) ensureChannel() error {
	if b.conn == nil || b.conn.IsClosed() {
		conn, err := amqp.Dial(b.uri)
		if err != nil {
			return fmt.Errorf("dial broker: %w", err)
		}
		b.conn = conn
		b.ch = nil
	}
	if b.ch == nil || b.ch.IsClosed() {
		ch, err := b.conn.Channel()
		if err != nil {
			return fmt.Errorf("open channel: %w", err)
		}
		b.ch = ch
		b.declared = make(map[string]bool)
	}
	return nil
}

func (b *broker) close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.conn == nil || b.conn.IsClosed() {
		return nil
	}
	return b.conn.Close() // also closes its channels
}
