package main

import "github.com/RadikChernyshov/get-social-web-service/pkg/queue"

// Start consuming, processing and aggregation of the messages (created by Web Service) from the Queue.
func main() {
	queue.Consume()
}
