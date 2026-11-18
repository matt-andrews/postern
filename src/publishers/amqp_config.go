package publishers

import (
	"github.com/osteele/liquid"
)

type AmqpConfig struct {
	uri     string
	topic   string
	message string
}

func (c AmqpConfig) Init(cfg map[string]string, bindings map[string]any) PublisherConfig {
	engine := liquid.NewEngine()

	c.uri = liquidRender(cfg["uri"], engine, bindings)
	c.topic = liquidRender(cfg["topic"], engine, bindings)
	c.message = liquidRender(cfg["message"], engine, bindings)
	return c
}
