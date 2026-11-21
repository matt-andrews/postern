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

func (c KafkaConfig) Init(cfg map[string]any, bindings map[string]any) interfaces.PublisherConfig {

	engine := liquid.NewEngine()

	c.uri = interfaces.LiquidRender(cfg["uri"].(string), engine, bindings)
	c.topic = interfaces.LiquidRender(cfg["topic"].(string), engine, bindings)
	c.message = interfaces.LiquidRender(cfg["message"].(string), engine, bindings)
	c.key = interfaces.LiquidRender(cfg["key"].(string), engine, bindings)

	return c
}
