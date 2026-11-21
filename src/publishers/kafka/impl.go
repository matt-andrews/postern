package kafka

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type WriterPool struct {
	mu      sync.RWMutex
	writers map[string]*kafka.Writer
}

func NewWriterPool() *WriterPool {
	return &WriterPool{
		writers: make(map[string]*kafka.Writer),
	}
}

// returns an existing writer or initializes a new one
func (p *WriterPool) getOrCreate(brokers []string, topic string) *kafka.Writer {
	// create cache key from broker and topic
	key := fmt.Sprintf("%s|%s", strings.Join(brokers, ","), topic)

	p.mu.RLock()
	w, exists := p.writers[key]
	p.mu.RUnlock()
	if exists {
		return w
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if w, exists = p.writers[key]; exists {
		return w
	}

	w = &kafka.Writer{
		AllowAutoTopicCreation: true,
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireAll,
		BatchTimeout:           10 * time.Millisecond,
	}
	p.writers[key] = w
	return w
}

func (p *WriterPool) Publish(ctx context.Context, brokers []string, topic string, msgs ...kafka.Message) error {
	w := p.getOrCreate(brokers, topic)
	return w.WriteMessages(ctx, msgs...)
}

func (p *WriterPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var firstErr error
	for key, w := range p.writers {
		if err := w.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(p.writers, key)
	}
	return firstErr
}
