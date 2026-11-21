package kafka

import (
	"context"
	"fmt"
	"matt-andrews/postern/publishers/interfaces"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	pool WriterPool
	ctx  context.Context
}

func NewKafkaPublisher() *KafkaPublisher {
	return &KafkaPublisher{pool: *NewWriterPool(), ctx: context.Background()}
}

func (p *KafkaPublisher) Key() string {
	return "kafka"
}

func (p *KafkaPublisher) Config(cfg map[string]any, bindings map[string]any) interfaces.PublisherConfig {
	return KafkaConfig{}.Init(cfg, bindings)
}

func (p *KafkaPublisher) Publish(config interfaces.PublisherConfig) error {
	cfg, ok := config.(KafkaConfig)
	if !ok {
		return fmt.Errorf("amqp publisher: expected AmqpConfig, got %T", config)
	}

	msg := kafka.Message{
		Value: []byte(cfg.message),
	}
	if cfg.key == "" {
		msg.Key = nil
	} else {
		msg.Key = []byte(cfg.key)
	}

	//check for headers and add them to the msg
	if cfg.headers != nil {
		for k, v := range cfg.headers {
			msg.Headers = append(msg.Headers, kafka.Header{
				Key:   k,
				Value: []byte(v),
			})
		}
	}

	return p.pool.Publish(p.ctx, cfg.brokers, cfg.topic, msg)
}

func (p *KafkaPublisher) Close() error {
	return p.pool.Close()
}
