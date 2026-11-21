package kafka

import (
	"matt-andrews/postern/publishers/interfaces"

	"github.com/osteele/liquid"
)

type KafkaConfig struct {
	brokers []string
	topic   string
	key     string
	message string
	headers map[string]string
}

func (c KafkaConfig) Init(cfg map[string]any, bindings map[string]any) interfaces.PublisherConfig {

	engine := liquid.NewEngine()

	c.brokers = interfaces.RenderStringArray(cfg["brokers"], engine, bindings)
	c.topic = interfaces.RenderString(cfg["topic"], engine, bindings)
	c.message = interfaces.RenderString(cfg["message"], engine, bindings)
	c.key = interfaces.RenderString(cfg["key"], engine, bindings)
	c.headers = interfaces.RenderStringMap(cfg["headers"], engine, bindings)

	return c
}
