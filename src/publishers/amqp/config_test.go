package amqp

import "testing"

// /ensure basic mapping without any liquid
func TestInitBasic(t *testing.T) {
	cfg := make(map[string]any)
	cfg["uri"] = "test-uri"
	cfg["topic"] = "test-topic"
	cfg["message"] = "test-message"
	bindings := make(map[string]any)
	want := AmqpConfig{
		uri:     "test-uri",
		topic:   "test-topic",
		message: "test-message",
	}

	result := initConfig(cfg, bindings, t)

	if result.uri != want.uri {
		t.Errorf("uri does not match")
	}
	if result.topic != want.topic {
		t.Errorf("topic does not match")
	}
	if result.message != want.message {
		t.Errorf("message does not match")
	}
}

// /test some basic liquid rendering on expected fields
func TestInitWithLiquid(t *testing.T) {
	cfg := make(map[string]any)
	cfg["uri"] = "{{ query.uri }}"
	cfg["topic"] = "{{ route.topic }}"
	cfg["message"] = "{{ body }}"
	bindings := make(map[string]any)
	bindings["query"] = map[string]string{
		"uri": "test-uri",
	}
	bindings["route"] = map[string]string{
		"topic": "test-topic",
	}
	bindings["body"] = "test-message"
	want := AmqpConfig{
		uri:     "test-uri",
		topic:   "test-topic",
		message: "test-message",
	}

	result := initConfig(cfg, bindings, t)

	if result.uri != want.uri {
		t.Errorf("uri does not match")
	}
	if result.topic != want.topic {
		t.Errorf("topic does not match")
	}
	if result.message != want.message {
		t.Errorf("message does not match")
	}
}

func initConfig(cfg map[string]any, bindings map[string]any, t *testing.T) AmqpConfig {
	init := AmqpConfig{}.Init(cfg, bindings)

	result, ok := init.(AmqpConfig)
	if !ok {
		t.Errorf("Config type was expected to be AmpqConfig")
	}
	return result
}
