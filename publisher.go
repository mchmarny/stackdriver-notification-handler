package main

import (
	"context"

	"cloud.google.com/go/pubsub"
)

func publish(ctx context.Context, data []byte) error {
	client, e := pubsub.NewClient(ctx, projectID)
	if e != nil {
		return e
	}
	topic := client.Topic(topicName)
	msg := &pubsub.Message{Data: data}
	result := topic.Publish(ctx, msg)
	_, err := result.Get(ctx)
	return err
}
