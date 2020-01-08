package storage

import (
	"fmt"
	"github.com/RadikChernyshov/get-social-web-service/pkg/logger"
	"github.com/amorist/mango"
	"github.com/amorist/mango/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"time"
)

// Get environment values to initiate the Storage package
// Open the single/reusable DB Engine connection
var (
	dbName    = os.Getenv("MONGO_DB")
	host      = os.Getenv("MONGO_HOST")
	port      = os.Getenv("MONGO_PORT")
	user      = os.Getenv("MONGO_USER")
	password  = os.Getenv("MONGO_PASSWORD")
	dbSession = mango.New(fmt.Sprintf("mongodb://%s:%s@%s:%s", user, password, host, port))
)

// Structure that represents input params for retrieving the data from the storage
type EventsQuery struct {
	Type     string
	Interval int
	From     int
	To       int
	Limit    int64
	Offset   int64
}

// Structure that represents input params that will be saved to storage
type EventSource struct {
	EventType string                 `bson:"eventType" json:"event_type"`
	Timestamp int                    `bson:"timestamp" json:"ts"`
	Params    map[string]interface{} `bson:"params" json:"params"`
}

// Structure that represents response from the storage engine
type EventResult struct {
	Id        primitive.ObjectID     `bson:"_id" json:"id"`
	EventType string                 `json:"eventType"`
	Timestamp int                    `json:"ts"`
	Params    map[string]interface{} `json:"params"`
}

// Initiate the Storage package and set the connection configuration values/params
func init() {
	dbSession.SetPoolLimit(10)
	if err := dbSession.Connect(); err != nil {
		logger.Fatal(err)
		return
	}
}

// Create/Save the new event data to storage. returns the error in case of issues during the saving process
func CreateEvent(event EventSource) error {
	return dbSession.DB(dbName).Collection("events").Insert(event)
}

// Retries the list of events events saved in storage
// Implements the filters mechanism to the DB Engine query to retrieve the data according to different input parameters
// Returns the error in case of issues during the select query or issues inside the DB Engine
func GetEvents(q *EventsQuery) (err error, result []EventResult) {
	condition := bson.M{}
	if q.Type != "" {
		condition["eventType"] = bson.M{"$eq": q.Type}
	}
	if q.From > 0 && q.To == 0 {
		condition["timestamp"] = bson.M{"$gte": q.From}
	}
	if q.To > 0 && q.From == 0 {
		condition["timestamp"] = bson.M{"$lte": q.To}
	}
	if q.To > 0 && q.From > 0 {
		condition["$and"] = []bson.M{
			{"timestamp": bson.M{"$gte": q.From}},
			{"timestamp": bson.M{"$lte": q.To}},
		}
	}
	if q.Interval > 0 {
		to := int32(time.Now().Unix())
		from := to - int32(q.Interval*3600)
		condition["$and"] = []bson.M{
			{"timestamp": bson.M{"$gte": from}},
			{"timestamp": bson.M{"$lte": to}},
		}
	}
	query := dbSession.DB(dbName).Collection("events").Find(condition)
	if q.Limit > 0 {
		query.Limit(q.Limit)
	} else {
		query.Limit(1000)
	}
	if q.Offset > 0 {
		query.Skip(q.Offset)
	}
	err = query.Sort(bson.M{"timestamp": 1}).All(&result)
	return err, result
}
