package amqp

import (
	"matt-andrews/postern/publishers/interfaces"

	"github.com/osteele/liquid"
)

type AmqpConfig struct {
	uri     string
	topic   string
	message string
}

func (c AmqpConfig) Init(cfg map[string]string, bindings map[string]any) interfaces.PublisherConfig {
	engine := liquid.NewEngine()

	c.uri = interfaces.LiquidRender(cfg["uri"], engine, bindings)
	c.topic = interfaces.LiquidRender(cfg["topic"], engine, bindings)
	c.message = interfaces.LiquidRender(cfg["message"], engine, bindings)
	return c
}
