/*
Package chaosmonkey provides a client to the Chaos Monkey REST API.

The client can be used to trigger chaos events, thereby causing Chaos Monkey to
"break" EC2 instances in different ways, to simulate different types of
failures:

	client, err := chaosmonkey.NewClient(&chaosmonkey.Config{
		Endpoint: "http://example.com:8080",
	})
	if err != nil {
		// handle error
	}
	event, err := client.TriggerEvent("ExampleAutoScalingGroup",
		chaosmonkey.StrategyShutdownInstance)
	...

Similar, the client can be used to retrieve information about past chaos events:

	events, err := client.Events()
	...

Note that in order to trigger chaos events, Chaos Monkey must be unleashed and
on-demand termination must be enabled via these configuration properties:

	simianarmy.chaos.leashed = false
	simianarmy.chaos.terminateOndemand.enabled = true
*/
package chaosmonkey

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// API constants
const (
	APIPath = "/simianarmy/api/v1/chaos"
)

// APIRequest describes a request sent to the API.
type APIRequest struct {
	ChaosType string `json:"chaosType,omitempty"`
	EventType string `json:"eventType"`
	GroupName string `json:"groupName"`
	GroupType string `json:"groupType"`
	Region    string `json:"region,omitempty"` // Ignored by vanilla Chaos Monkey
}

// APIResponse describes a response returned by the API.
type APIResponse struct {
	ChaosType  string `json:"chaosType,omitempty"`
	EventID    string `json:"eventId"`
	EventTime  int64  `json:"eventTime"`
	EventType  string `json:"eventType"`
	GroupName  string `json:"groupName"`
	GroupType  string `json:"groupType"`
	MonkeyType string `json:"monkeyType"`
	Region     string `json:"region"`
	Message    string `json:"message,omitempty"`
}

// ToEvent converts the APIResponse into an Event.
func (resp *APIResponse) ToEvent() *Event {
	return &Event{
		InstanceID:           resp.EventID,
		AutoScalingGroupName: resp.GroupName,
		Region:               resp.Region,
		Strategy:             Strategy(resp.ChaosType),
		TriggeredAt:          time.Unix(resp.EventTime/1000, 0).UTC(),
	}
}

// Event describes the termination of an EC2 instance by Chaos Monkey.
type Event struct {
	// ID of EC2 instance that was terminated
	InstanceID string

	// Name of auto scaling group containing the terminated instance
	AutoScalingGroupName string

	// AWS region of the instance and its auto scaling group
	Region string

	// Chaos strategy used to terminate the instance
	Strategy Strategy

	// Time when the chaos event was triggered
	TriggeredAt time.Time
}

// Config is used to configure the creation of the client.
type Config struct {
	// Address and port of the Chaos Monkey API server
	Endpoint string

	// Optional AWS region (ignored by vanilla Chaos Monkey)
	Region string

	// Optional username for HTTP Basic Authentication
	Username string

	// Optional password for HTTP Basic Authentication
	Password string

	// Custom HTTP User Agent
	UserAgent string

	// Custom HTTP client to use (http.DefaultClient by default)
	HTTPClient *http.Client
}

// DefaultConfig returns a default configuration for the client. It parses the
// environment variables CHAOSMONKEY_ENDPOINT, CHAOSMONKEY_USERNAME, and
// CHAOSMONKEY_PASSWORD.
func DefaultConfig() *Config {
	c := Config{
		Endpoint:   "http://127.0.0.1:8080",
		UserAgent:  "chaosmonkey Go library",
		HTTPClient: http.DefaultClient,
	}
	if v := os.Getenv("CHAOSMONKEY_ENDPOINT"); v != "" {
		c.Endpoint = v
	}
	if v := os.Getenv("CHAOSMONKEY_USERNAME"); v != "" {
		c.Username = v
	}
	if v := os.Getenv("CHAOSMONKEY_PASSWORD"); v != "" {
		c.Password = v
	}
	return &c
}

// Client is the client to the Chaos Monkey API. Create a client with NewClient.
type Client struct {
	config *Config
}

// NewClient returns a new client for the given configuration.
func NewClient(c *Config) (*Client, error) {
	defConfig := DefaultConfig()
	if c.Endpoint == "" {
		c.Endpoint = defConfig.Endpoint
	}
	if !strings.HasPrefix(c.Endpoint, "http") {
		c.Endpoint = "http://" + c.Endpoint
	}
	if c.Username == "" {
		c.Username = defConfig.Username
	}
	if c.Password == "" {
		c.Password = defConfig.Password
	}
	if c.UserAgent == "" {
		c.UserAgent = defConfig.UserAgent
	}
	if c.HTTPClient == nil {
		c.HTTPClient = defConfig.HTTPClient
	}
	return &Client{config: c}, nil
}

// TriggerEvent triggers a new chaos event which will cause Chaos Monkey to
// "break" an EC2 instance in the given auto scaling group using the specified
// chaos strategy.
func (c *Client) TriggerEvent(group string, strategy Strategy) (*Event, error) {
	url := c.config.Endpoint + APIPath

	body, err := json.Marshal(APIRequest{
		EventType: "CHAOS_TERMINATION",
		GroupType: "ASG",
		GroupName: group,
		ChaosType: string(strategy),
		Region:    c.config.Region,
	})
	if err != nil {
		return nil, err
	}

	var resp APIResponse
	if err := c.sendRequest("POST", url, bytes.NewReader(body), &resp); err != nil {
		return nil, err
	}

	return resp.ToEvent(), nil
}

// Events returns a list of all chaos events.
func (c *Client) Events() ([]Event, error) {
	return c.events(0)
}

// EventsSince returns a list of all chaos events since a specific time.
func (c *Client) EventsSince(t time.Time) ([]Event, error) {
	return c.events(t.UTC().Unix() * 1000)
}

func (c *Client) events(since int64) ([]Event, error) {
	url := fmt.Sprintf("%s%s?since=%d", c.config.Endpoint, APIPath, since)

	var resp []APIResponse
	if err := c.sendRequest("GET", url, nil, &resp); err != nil {
		return nil, err
	}

	var events []Event
	for _, r := range resp {
		events = append(events, *r.ToEvent())
	}

	return events, nil
}

func (c *Client) sendRequest(method, url string, body io.Reader, out interface{}) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	if c.config.Username != "" && c.config.Password != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}
	req.Header.Add("User-Agent", c.config.UserAgent)

	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return decodeError(resp)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func decodeError(resp *http.Response) error {
	var r APIResponse
	err := json.NewDecoder(resp.Body).Decode(&r)
	if err == nil && r.Message != "" {
		return fmt.Errorf("%s", r.Message)
	}
	return fmt.Errorf("HTTP error: %s", resp.Status)
}
