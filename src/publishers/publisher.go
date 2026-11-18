package publishers

import (
	"errors"
	"fmt"
)

type Publisher interface {
	Key() string
	Publish(cfg map[string]string) error
}

var registry = map[string]Publisher{}

func register(p Publisher) { registry[p.Key()] = p }

func Init() {
	register(NewAmqpPublisher())
}

func PublishAll(outputs []map[string]string) error {
	var errs []error
	for _, out := range outputs {
		pub, ok := registry[out["type"]]
		if !ok {
			errs = append(errs, fmt.Errorf("unknown publisher type %q", out["type"]))
			continue
		}
		if err := pub.Publish(out); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
