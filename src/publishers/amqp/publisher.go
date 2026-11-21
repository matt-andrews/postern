package amqp

import (
	"context"
	"errors"
	"fmt"
	"matt-andrews/postern/publishers/interfaces"
	"sync"
	"time"
)

type AmqpPublisher struct {
	mu      sync.Mutex
	brokers map[string]*broker
	dial    dialFunc
}

func NewAmqpPublisher() *AmqpPublisher {
	return &AmqpPublisher{brokers: make(map[string]*broker), dial: dialAMQP}
}

func (p *AmqpPublisher) Key() string {
	return "amqp"
}

func (p *AmqpPublisher) Config(cfg map[string]any, bindings map[string]any) interfaces.PublisherConfig {
	return AmqpConfig{}.Init(cfg, bindings)
}

func (p *AmqpPublisher) Publish(config interfaces.PublisherConfig) error {
	cfg, ok := config.(AmqpConfig)
	if !ok {
		return fmt.Errorf("amqp publisher: expected AmqpConfig, got %T", config)
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
		b = &broker{uri: uri, dial: p.dial}
		p.brokers[uri] = b
	}
	return b
}
