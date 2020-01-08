package api

import (
	"encoding/json"
	"fmt"
	"github.com/RadikChernyshov/get-social-web-service/pkg/environment"
	"github.com/RadikChernyshov/get-social-web-service/pkg/logger"
	"github.com/RadikChernyshov/get-social-web-service/pkg/queue"
	"github.com/RadikChernyshov/get-social-web-service/pkg/storage"
	"github.com/valyala/fasthttp"
	"strconv"
	"time"
)

// Representation structure for the new event that should be sent
// to the Queue for further processing
type CreateEventPayload struct {
	EventType string                 `json:"event_type"`
	Timestamp int                    `json:"ts"`
	Params    map[string]interface{} `json:"params"`
}

// Representation structure for the HTTP response after a new event
// is successfully created (sent to the Queue)
type CreateEventResponse struct {
	Status int                 `json:"status"`
	Data   *CreateEventPayload `json:"data"`
}

// Representation structure for the HTTP response on getting existing events
type GetEventsResponse struct {
	Status int                   `json:"status"`
	Data   []storage.EventResult `json:"data"`
}

// Representation structure for the HTTP response to get the web server status
type HealthCheckResponse struct {
	Status    int   `json:"status"`
	Timestamp int64 `json:"timestamp"`
}

// Representation structure for the HTTP response when the requested page by client
// was not found or the route was not correctly requested
type NotFoundResponse struct {
	Timestamp int64 `json:"timestamp"`
}

// Representation structure for the HTTP response when the requested payload is invalid
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Standardize the way how the HTTP response should be sent to the API client.
// Logs the requested url to stdout in development mode.
func responseJson(ctx *fasthttp.RequestCtx, responseStatusCode int, responseBody interface{}) {
	if environment.Development() {
		logger.Info(fmt.Sprintf("%q %q %q", ctx.Method(), ctx.Host(), ctx.RequestURI()))
	}
	ctx.Response.Header.SetCanonical([]byte("Content-Type"), []byte("application/json"))
	ctx.Response.SetStatusCode(responseStatusCode)
	err := json.NewEncoder(ctx).Encode(responseBody)
	if err != nil {
		logger.Warning(err.Error())
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
}

// Functional handler for the requests that was to routed incorrectly
func NotFound(ctx *fasthttp.RequestCtx) {
	responseBody := ErrorResponse{
		Code:    fasthttp.StatusNotFound,
		Message: "requested resource not found",
	}
	responseJson(ctx, fasthttp.StatusNotFound, responseBody)
}

// Functional handler for the healthcheck requests
func IndexGet(ctx *fasthttp.RequestCtx) {
	responseBody := HealthCheckResponse{
		Status:    fasthttp.StatusOK,
		Timestamp: time.Now().Unix(),
	}
	responseJson(ctx, fasthttp.StatusOK, responseBody)
}

// Functional handler for create events route
func EventsCreate(ctx *fasthttp.RequestCtx) {
	eventPayload := new(CreateEventPayload)
	requestBody := ctx.PostBody()
	err := json.Unmarshal(requestBody, &eventPayload)
	if err != nil {
		responseJson(ctx, fasthttp.StatusUnprocessableEntity, err)
		return
	}
	if eventPayload.EventType == "" {
		responseJson(ctx, fasthttp.StatusUnprocessableEntity, ErrorResponse{fasthttp.StatusUnprocessableEntity, "event has to have a type"})
		return
	}
	if eventPayload.Timestamp == 0 {
		responseJson(ctx, fasthttp.StatusUnprocessableEntity, ErrorResponse{fasthttp.StatusUnprocessableEntity, "event has to have a timestamp"})
		return
	}
	if published := queue.Publish(eventPayload); !published {
		logger.Warning("event has not been sent to queue")
		responseJson(ctx, fasthttp.StatusUnprocessableEntity, ErrorResponse{fasthttp.StatusUnprocessableEntity, "event has not been processed"})
		return
	}
	responseBody := CreateEventResponse{
		Status: fasthttp.StatusCreated,
		Data:   eventPayload,
	}
	responseJson(ctx, fasthttp.StatusCreated, responseBody)
}

// Functional handler for get events route
func EventsGet(ctx *fasthttp.RequestCtx) {
	from, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("from")))
	to, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("to")))
	limit, _ := strconv.ParseInt(string(ctx.QueryArgs().Peek("limit")), 10, 64)
	offset, _ := strconv.ParseInt(string(ctx.QueryArgs().Peek("offset")), 10, 64)
	interval, _ := strconv.Atoi(string(ctx.QueryArgs().Peek("interval")))
	eventQuery := new(storage.EventsQuery)
	eventQuery.Type = string(ctx.QueryArgs().Peek("type"))
	eventQuery.From = from
	eventQuery.Interval = interval
	eventQuery.To = to
	eventQuery.Limit = limit
	eventQuery.Offset = offset
	err, events := storage.GetEvents(eventQuery)
	if err != nil {
		responseJson(ctx, fasthttp.StatusUnprocessableEntity, err)
		return
	}
	responseBody := GetEventsResponse{
		Status: fasthttp.StatusOK,
		Data:   events,
	}

	responseJson(ctx, fasthttp.StatusOK, responseBody)
}
