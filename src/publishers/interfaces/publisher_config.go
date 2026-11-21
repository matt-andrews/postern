package interfaces

import (
	"fmt"
	"log"

	"github.com/osteele/liquid"
)

type PublisherConfig interface {
	Init(cfg map[string]any, bindings map[string]any) PublisherConfig
}

func liquidRender(in string, engine *liquid.Engine, bindings map[string]any) string {
	out, err := engine.ParseAndRenderString(in, bindings)
	if err != nil {
		log.Println(err)
		return in
	}
	return out
}

func RenderString(in any, engine *liquid.Engine, bindings map[string]any) string {
	switch v := in.(type) {
	case nil:
		return ""
	case string:
		return liquidRender(v, engine, bindings)
	default:
		return liquidRender(fmt.Sprint(v), engine, bindings)
	}
}

func RenderStringArray(in any, engine *liquid.Engine, bindings map[string]any) []string {
	var items []string

	switch v := in.(type) {
	case nil:
		return nil
	case string:
		items = []string{v}
	case []string:
		items = v
	case []any:
		items = make([]string, 0, len(v))
		for _, e := range v {
			s, ok := e.(string)
			if !ok {
				s = fmt.Sprint(e)
			}
			items = append(items, s)
		}
	default:
		items = []string{fmt.Sprint(v)}
	}

	result := make([]string, len(items))
	for i, item := range items {
		result[i] = liquidRender(item, engine, bindings)
	}
	return result
}

func RenderStringMap(in any, engine *liquid.Engine, bindings map[string]any) map[string]string {
	var items map[string]any

	switch v := in.(type) {
	case nil:
		return nil
	case map[string]any:
		items = v
	case map[any]any: // handle yaml.v2
		items = make(map[string]any, len(v))
		for k, val := range v {
			items[fmt.Sprint(k)] = val
		}
	default:
		return nil
	}

	result := make(map[string]string, len(items))
	for k, val := range items {
		s, ok := val.(string)
		if !ok {
			s = fmt.Sprint(val)
		}
		result[k] = liquidRender(s, engine, bindings)
	}
	return result
}
