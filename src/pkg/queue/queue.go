package queue

import (
	"encoding/json"
	"fmt"
	"github.com/RadikChernyshov/get-social-web-service/pkg/logger"
	"github.com/RadikChernyshov/get-social-web-service/pkg/storage"
	"github.com/adjust/rmq"
	"os"
	"strconv"
	"time"
)

// Get environment values to initiate the Queue package
// Open the single/reusable Queue Storage Engine connection
var (
	host            = os.Getenv("REDIS_HOST")
	port            = os.Getenv("REDIS_PORT")
	connection      = rmq.OpenConnection("events", "tcp", fmt.Sprintf("%v:%v", host, port), 1)
	unackedLimit    = 20
	numConsumers    = 2
	activeQueueName = "events"
)

// Structure that represents input params that will be sent to Queue storage
type TackPayload struct {
	EventType string                 `json:"event_type"`
	Timestamp int                    `json:"ts"`
	Params    map[string]interface{} `json:"params"`
}

// Structure that represents worker mechanism the process the messages from the queue
type Consumer struct {
	name   string
	count  int
	before time.Time
}

// Initiate the package and retrieve the values/params from the process environment
func init() {
	numConsumers, err := strconv.Atoi(os.Getenv("CONSUMERS_COUNT"))
	if err != nil || numConsumers == 0 {
		numConsumers = 1
	}
}

// Save/Publish message to the queue engine return positive value in case of success
// and negative in case Queue Storage Engine of issues
func Publish(payload interface{}) bool {
	taskQueue := connection.OpenQueue(activeQueueName)
	taskBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Warning(err)
		return false
	}
	return taskQueue.PublishBytes(taskBytes)
}

// Process the messages that was sent to the queue by Publish method.
// Initiate and configure the consumers processes
func Consume() {
	queue := connection.OpenQueue(activeQueueName)
	queue.StartConsuming(unackedLimit, 500*time.Millisecond)
	for i := 0; i < numConsumers; i++ {
		name := fmt.Sprintf("Consumer %d", i)
		queue.AddConsumer(name, NewConsumer(i))
	}
	select {}
}

// Initiate and configure the new consumers processes
func NewConsumer(tag int) *Consumer {
	return &Consumer{
		name:   fmt.Sprintf("Consumer%d", tag),
		count:  0,
		before: time.Now(),
	}
}

// Logic that implements the behavior of the consumer process
// Rejects the message in case of invalid queue message
// Rejects the message in case of issues in Storage Engine
// Logs the process errors during message processing
// Accepts the message in case of success
func (consumer *Consumer) Consume(delivery rmq.Delivery) {
	var event storage.EventSource
	if err := json.Unmarshal([]byte(delivery.Payload()), &event); err != nil {
		logger.Warning("event error: %s", err)
		delivery.Reject()
		return
	}
	if err := storage.CreateEvent(event); err != nil {
		logger.Warning("create record error: %s", err)
		delivery.Reject()
		return
	}
	delivery.Ack()
}
