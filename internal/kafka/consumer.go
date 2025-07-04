package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"wbtech"
	"wbtech/internal/service"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader  *kafka.Reader
	service *service.OrderService
}

func NewConsumer(brokers []string, topic string, groupID string, s *service.OrderService) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
		service: s,
	}
}

func (c *Consumer) Consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)

			if errors.Is(err, context.Canceled) {
				slog.Info("kafka consumer stopped gracefully")
				return
			} else if err != nil {
				slog.Error("kafka consume error", "error", err)
			}

			var order wbtech.Order

			if err := json.Unmarshal(msg.Value, &order); err != nil {
				slog.Error(fmt.Sprintf("failed to unmarshal order: %v", err))
				continue
			}

			if err := c.service.ProcessOrder(ctx, order); err != nil {
				slog.Error(fmt.Sprintf("failed to process order: %v", err))
				continue
			}

			slog.Info(fmt.Sprintf("successfully processed order: %s", order.OrderUID))
		}

	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
