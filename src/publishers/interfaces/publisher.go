package interfaces

type Publisher interface {
	Key() string
	Publish(cfg PublisherConfig) error
	Config(cfg map[string]string, bindings map[string]any) PublisherConfig
}
