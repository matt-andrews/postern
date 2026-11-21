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

func (c AmqpConfig) Init(cfg map[string]any, bindings map[string]any) interfaces.PublisherConfig {
	engine := liquid.NewEngine()

	c.uri = interfaces.RenderString(cfg["uri"], engine, bindings)
	c.topic = interfaces.RenderString(cfg["topic"], engine, bindings)
	c.message = interfaces.RenderString(cfg["message"], engine, bindings)
	return c
}
