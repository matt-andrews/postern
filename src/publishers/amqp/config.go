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
	if cfg["message"] == nil || cfg["message"] == "" {
		if body, ok := bindings["body"]; ok {
			c.message = body.(string)
		}
	} else {
		c.message = interfaces.RenderString(cfg["message"], engine, bindings)
	}
	return c
}
