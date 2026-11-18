package publishers

import (
	"errors"
	"fmt"
	"matt-andrews/postern/publishers/amqp"
	"matt-andrews/postern/publishers/interfaces"
)

var registry = map[string]interfaces.Publisher{}

func register(p interfaces.Publisher) { registry[p.Key()] = p }

func Init() {
	register(amqp.NewAmqpPublisher())
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
