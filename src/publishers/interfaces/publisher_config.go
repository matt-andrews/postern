package interfaces

import (
	"log"

	"github.com/osteele/liquid"
)

type PublisherConfig interface {
	Init(cfg map[string]string, bindings map[string]any) PublisherConfig
}

func LiquidRender(in string, engine *liquid.Engine, bindings map[string]any) string {
	out, err := engine.ParseAndRenderString(in, bindings)
	if err != nil {
		log.Println(err)
		return in
	}
	return out
}
