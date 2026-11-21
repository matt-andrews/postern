package interfaces

type Publisher interface {
	Key() string
	Publish(cfg PublisherConfig) error
	Config(cfg map[string]any, bindings map[string]any) PublisherConfig
	Close() error
}
