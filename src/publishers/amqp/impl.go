package amqp

import (
	"context"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type amqpConn interface {
	Channel() (amqpChannel, error)
	IsClosed() bool
	Close() error
}

type amqpChannel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	IsClosed() bool
}

type dialFunc func(uri string) (amqpConn, error)
type realConn struct{ *amqp.Connection }

func (c realConn) Channel() (amqpChannel, error) {
	ch, err := c.Connection.Channel()
	if err != nil {
		return nil, err
	}
	return ch, nil
}
func dialAMQP(uri string) (amqpConn, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}
	return realConn{conn}, nil
}

type broker struct {
	uri  string
	dial dialFunc

	mu       sync.Mutex
	conn     amqpConn
	ch       amqpChannel
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
		conn, err := b.dial(b.uri)
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
