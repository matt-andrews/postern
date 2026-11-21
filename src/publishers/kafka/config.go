package kafka

import (
	"matt-andrews/postern/publishers/interfaces"

	"github.com/osteele/liquid"
)

type KafkaConfig struct {
	uri     string
	topic   string
	key     string
	message string
}

func (c KafkaConfig) Init(cfg map[string]string, bindings map[string]any) interfaces.PublisherConfig {

	engine := liquid.NewEngine()

	c.uri = interfaces.LiquidRender(cfg["uri"], engine, bindings)
	c.topic = interfaces.LiquidRender(cfg["topic"], engine, bindings)
	c.message = interfaces.LiquidRender(cfg["message"], engine, bindings)
	c.key = interfaces.LiquidRender(cfg["key"], engine, bindings)

	return c
}
