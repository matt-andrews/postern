package publishers

import (
	"matt-andrews/postern/publishers/interfaces"
	"testing"
)

// -----mocks-----//
type fakePublisher struct {
}

func (f fakePublisher) Key() string {
	return "fake"
}
func (f fakePublisher) Publish(cfg interfaces.PublisherConfig) error {
	return nil
}
func (f fakePublisher) Config(cfg map[string]any, bindings map[string]any) interfaces.PublisherConfig {
	return fakeConfig{}
}
func (f fakePublisher) Close() error {
	return nil
}

type fakeConfig struct{}

func (f fakeConfig) Init(cfg map[string]any, bindings map[string]any) interfaces.PublisherConfig {
	return fakeConfig{}
}

//-----tests-----//

func TestFactory_PublishAll_Success(t *testing.T) {
	register(fakePublisher{})

	bindings := make(map[string]any)
	tests := []map[string]any{
		{
			"type": "fake",
		},
	}

	err := PublishAll(tests, bindings)
	if err != nil {
		t.Errorf("failed to publish at least one: %e", err)
	}
}

func TestFactory_PublishAll_FailedToFind(t *testing.T) {
	register(fakePublisher{})

	bindings := make(map[string]any)
	tests := []map[string]any{
		{
			"type": "fake-cant-find",
		},
	}

	err := PublishAll(tests, bindings)
	if err == nil {
		t.Errorf("failed to error out when not finding a publisher")
	}
}
