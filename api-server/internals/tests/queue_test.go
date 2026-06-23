package tests

import (
	"context"
	"testing"

	"github.com/joho/godotenv"
	queue "github.com/uddinArsalan/devdeploy/internals/adapters/messenger"
	"github.com/uddinArsalan/devdeploy/internals/domain"
)

func TestQueue(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatal(err)
	}
	rmqClient, err := queue.NewRabbitMQClient(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	err = rmqClient.PublishMessage(context.Background(), domain.BuildJob{
		ProjectID: 1,
		GitURL:    "hello.github",
		DeployID:  12,
		Slug:      "hello-slug",
	})
	if err != nil {
		t.Fatalf("error publishing message %v", err)
	}
	for {
		msg, err := rmqClient.ConsumeMessage(context.Background())
		if err != nil {
			t.Fatalf("error consuming message %v", err)
		}
		t.Logf("Recieved message %v", msg)
		t.Logf("Recieved message %v", msg)
	}
}
