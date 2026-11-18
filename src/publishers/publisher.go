package publishers

import (
	"errors"
	"fmt"
)

type Publisher interface {
	Key() string
	Publish(cfg PublisherConfig) error
	Config(cfg map[string]string, bindings map[string]any) PublisherConfig
}

var registry = map[string]Publisher{}

func register(p Publisher) { registry[p.Key()] = p }

func Init() {
	register(NewAmqpPublisher())
}

func PublishAll(outputs []map[string]string, bindings map[string]any) error {
	var errs []error
	for _, out := range outputs {
		pub, ok := registry[out["type"]]
		if !ok {
			errs = append(errs, fmt.Errorf("unknown publisher type %q", out["type"]))
			continue
		}

		cfg := pub.Config(out, bindings)
		if err := pub.Publish(cfg); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
