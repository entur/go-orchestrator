package orchestrator

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/cloudevents/sdk-go/v2/event"
)

func ParseEvent[T any](e event.Event) (Request[T], error) {
	var req Request[T]
	var data EventData
	err := e.DataAs(&data)
	if err != nil {
		return req, err
	}

	err = json.Unmarshal(data.Message.Data, &req)
	if err != nil {
		return req, err
	}
	return req, nil
}

type PubSubMessageAttributes struct{}

type PubSubMessage struct {
	ID          string                  `json:"messageId"`
	PublishTime string                  `json:"publishTime"`
	Attributes  PubSubMessageAttributes `json:"attributes"`
	Data        []byte                  `json:"data"`
}

type EventData struct {
	Subscription string
	Message      PubSubMessage
}

type TopicCache struct {
	mu     sync.Mutex
	client *pubsub.Client
	topics map[string]*pubsub.Topic
}

func (c *TopicCache) Topics() []*pubsub.Topic {
	c.mu.Lock()
	defer c.mu.Unlock()

	var topics []*pubsub.Topic

	num := len(c.topics)
	if num > 0 {
		topics := make([]string, 0, num)
		for _, topic := range topics {
			topics = append(topics, topic)
		}
	}

	return topics
}

func (c *TopicCache) Topic(projectID string, topicID string) *pubsub.Topic {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := projectID + topicID

	topic, ok := c.topics[key]
	if !ok {
		topic = c.client.TopicInProject(topicID, projectID)
		c.topics[key] = topic
	}

	return topic
}

func (c *TopicCache) TopicFullID(id string) *pubsub.Topic {
	if !strings.HasPrefix(id, "projects/") {
		return nil
	}

	i := strings.Index(id[9:], "/")
	if i == -1 {
		return nil
	}

	projectID := id[9 : 9+i]
	topicID := id[strings.LastIndex(id, "/")+1:]

	return c.Topic(projectID, topicID)
}

func NewTopicCache(client *pubsub.Client) *TopicCache {
	return &TopicCache{
		client: client,
		topics: map[string]*pubsub.Topic{},
	}
}

func NewMockEvent[T any](manifest T, sender SenderType, action Action) (*event.Event, error) {
	req := Request[T]{
		ApiVersion: "orchestrator.entur.io/request/v1",
		Metadata:   OuterMetadata{},
		Sender: Sender{
			Type: sender,
		},
		Action:        action,
		ResponseTopic: "topic",
		Manifest:      Manifests[T]{Old: nil, New: manifest},
	}
	reqdata, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(reqdata)))
	base64.StdEncoding.Encode(buf, reqdata)
	data, err := json.Marshal(&EventData{
		Message: PubSubMessage{
			Data:        reqdata,
			ID:          "id",
			PublishTime: "time",
			Attributes:  PubSubMessageAttributes{},
		},
		Subscription: "sub",
	})
	if err != nil {
		return nil, err
	}

	event := event.New(event.CloudEventsVersionV03)
	event.DataEncoded = data
	return &event, nil
}
