package harness

import (
	"context"
	"sync"
	"testing"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/testcontainers/testcontainers-go"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

const kafkaPartitions = 3

type Kafka struct {
	Brokers []string
	ctx     context.Context
	mu      sync.Mutex
	feeds   map[string]chan kafkago.Message
}

type KafkaMessage struct {
	Key       string
	Body      string
	Headers   map[string]string
	Partition int
}

// starts a single-node Kafka container and stops it when the test ends.
func StartKafka(t testing.TB) *Kafka {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	ctr, err := tckafka.Run(ctx, "confluentinc/confluent-local:7.5.0",
		testcontainers.WithEnv(map[string]string{"KAFKA_NUM_PARTITIONS": "3"}),
	)
	// nothing in the container is worth a graceful shutdown, so kill it straight away
	testcontainers.CleanupContainer(t, ctr, testcontainers.StopTimeout(0))
	if err != nil {
		t.Fatalf("start kafka: %v", err)
	}

	brokers, err := ctr.Brokers(ctx)
	if err != nil {
		t.Fatalf("kafka brokers: %v", err)
	}

	return &Kafka{Brokers: brokers, ctx: ctx, feeds: make(map[string]chan kafkago.Message)}
}

// creates a uniquely named topic and returns its name.
// real clusters usually have topics created up front, so most stories should start here.
func (k *Kafka) CreateTopic(t testing.TB, prefix string) string {
	t.Helper()
	topic := UniqueName(prefix)

	conn, err := kafkago.DialContext(k.ctx, "tcp", k.Brokers[0])
	if err != nil {
		t.Fatalf("dial kafka: %v", err)
	}
	defer conn.Close()

	err = conn.CreateTopics(kafkago.TopicConfig{
		Topic:             topic,
		NumPartitions:     kafkaPartitions,
		ReplicationFactor: 1,
	})
	if err != nil {
		t.Fatalf("create topic %q: %v", topic, err)
	}
	return topic
}

// waits for the next message on topic and returns it.
// calling it again on the same topic returns the message after that.
func (k *Kafka) ExpectMessage(t testing.TB, topic string, timeout time.Duration) KafkaMessage {
	t.Helper()

	select {
	case m := <-k.feed(t, topic):
		headers := make(map[string]string, len(m.Headers))
		for _, h := range m.Headers {
			headers[h.Key] = string(h.Value)
		}
		return KafkaMessage{
			Key:       string(m.Key),
			Body:      string(m.Value),
			Headers:   headers,
			Partition: m.Partition,
		}
	case <-time.After(timeout):
		t.Fatalf("no message on %q after %s", topic, timeout)
		return KafkaMessage{}
	}
}

// returns a Go channel carrying every message on topic, oldest first per partition.
// the first call for a topic starts one background reader per partition;
// each begins at the start of its partition, so it does not matter that
// postern published before anyone was listening.
func (k *Kafka) feed(t testing.TB, topic string) chan kafkago.Message {
	t.Helper()
	k.mu.Lock()
	defer k.mu.Unlock()

	if feed, ok := k.feeds[topic]; ok {
		return feed
	}

	partitions, err := kafkago.LookupPartitions(k.ctx, "tcp", k.Brokers[0], topic)
	if err != nil {
		t.Fatalf("look up topic %q: %v", topic, err)
	}

	feed := make(chan kafkago.Message, 100)
	for _, p := range partitions {
		r := kafkago.NewReader(kafkago.ReaderConfig{
			Brokers:   k.Brokers,
			Topic:     topic,
			Partition: p.ID,
			MaxWait:   250 * time.Millisecond,
		})
		t.Cleanup(func() { r.Close() })

		go func() {
			for {
				m, err := r.ReadMessage(k.ctx)
				if err != nil {
					return // the test is over
				}
				feed <- m
			}
		}()
	}

	k.feeds[topic] = feed
	return feed
}
