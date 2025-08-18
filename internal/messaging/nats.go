package messaging

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/stan.go"
)

type NATSClient struct {
	conn stan.Conn
}

type Config struct {
	URL       string
	ClusterID string
	ClientID  string
}

func NewNATSClient(cfg Config) (*NATSClient, error) {
	// // Generate unique client ID to avoid conflicts
	// uniqueClientID := fmt.Sprintf("%s-%s", cfg.ClientID, uuid.New().String()[:8])

	// conn, err := stan.Connect(cfg.ClusterID, uniqueClientID, stan.NatsURL(cfg.URL))
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to connect to NATS Streaming: %w", err)
	// }

	// log.Printf("Connected to NATS Streaming: %s (cluster: %s, client: %s)",
	// 	cfg.URL, cfg.ClusterID, uniqueClientID)

	// return &NATSClient{conn: conn}, nil
	return &NATSClient{}, nil
}

func (nc *NATSClient) Publish(subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = nc.conn.Publish(subject, payload)
	if err != nil {
		return fmt.Errorf("failed to publish to subject %s: %w", subject, err)
	}

	log.Printf("Published message to subject: %s", subject)
	return nil
}

func (nc *NATSClient) Subscribe(subject string, handler stan.MsgHandler) (stan.Subscription, error) {
	sub, err := nc.conn.Subscribe(subject, handler, stan.DurableName(subject+"-durable"))
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to subject %s: %w", subject, err)
	}

	log.Printf("Subscribed to subject: %s", subject)
	return sub, nil
}

func (nc *NATSClient) SubscribeQueue(subject, queue string, handler stan.MsgHandler) (stan.Subscription, error) {
	sub, err := nc.conn.QueueSubscribe(subject, queue, handler,
		stan.DurableName(subject+"-"+queue+"-durable"),
		stan.AckWait(30*time.Second),
		stan.MaxInflight(1))
	if err != nil {
		return nil, fmt.Errorf("failed to queue subscribe to subject %s: %w", subject, err)
	}

	log.Printf("Subscribed to subject: %s (queue: %s)", subject, queue)
	return sub, nil
}

func (nc *NATSClient) Close() error {
	if nc.conn != nil {
		return nc.conn.Close()
	}
	return nil
}
