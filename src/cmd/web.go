package main

import (
	"flag"
	"fmt"
	"github.com/RadikChernyshov/get-social-web-service/pkg/api"
	"github.com/RadikChernyshov/get-social-web-service/pkg/logger"
)

var (
	addr = flag.String("addr", ":80", "TCP address to listen to")
)

// Start web server to accept REST calls.
// Fails if the web address/port is not available or can not be started.
func main() {
	if err := api.New(addr); err != nil {
		logger.Fatal(fmt.Sprintf("web server error: %s", err))
	}
}
