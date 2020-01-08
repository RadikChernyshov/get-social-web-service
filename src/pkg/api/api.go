package api

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

// Initialize the Web Service and it's routing returns error in case of issue with address acceptability.
func New(addr *string) error {
	r := router.New()
	r.NotFound = NotFound
	r.GET("/", IndexGet)
	r.GET("/api/v1/events", EventsGet)
	r.POST("/api/v1/events", EventsCreate)
	return fasthttp.ListenAndServe(*addr, r.Handler)
}
