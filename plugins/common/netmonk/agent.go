package netmonk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/influxdata/telegraf"
)

const (
	verifyEndpoint = "/public/controller/server/%s/verify"
)

// Agent common to authenticate netmonk agent.
type Agent struct {
	NetmonkHost      string `toml:"netmonk_host"`
	NetmonkServerID  string `toml:"netmonk_server_id"`
	NetmonkServerKey string `toml:"netmonk_server_key"`

	Log telegraf.Logger `toml:"-"`
}

type CustomerCredentials struct {
	ClientID      string        `json:"client_id"`
	MessageBroker MessageBroker `json:"message_broker"`
	Auth          Auth          `json:"auth"`
	SASL          SASL          `json:"sasl"`
	TLS           TLS           `json:"tls"`
}

type MessageBroker struct {
	Type      string   `json:"type"`
	Addresses []string `json:"address"`
}

type Auth struct {
	IsEnabled bool   `json:"is_enabled"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type SASL struct {
	IsEnabled bool   `json:"is_enabled"`
	Mechanism string `json:"mechanism"`
}

type TLS struct {
	IsEnabled bool   `json:"is_enabled"`
	CA        string `json:"ca"`
	Access    string `json:"access"`
	Key       string `json:"key"`
}

func NewAgent(host, serverid, serverkey string) *Agent {
	return &Agent{
		NetmonkHost:      host,
		NetmonkServerID:  serverid,
		NetmonkServerKey: serverkey,
	}
}

// ResponseFormat stands for our default response in API
type ResponseFormat struct {
	Data     interface{} `json:"data,omitempty"`
	Errors   interface{} `json:"errors,omitempty"`
	Meta     interface{} `json:"meta,omitempty"`
	Jsonapi  interface{} `json:"jsonapi,omitempty"`
	Links    interface{} `json:"links,omitempty"`
	Included interface{} `json:"included,omitempty"`
}

// Verify netmonk agent.
func (n *Agent) Verify() (*CustomerCredentials, error) {
	postBody, _ := json.Marshal(map[string]string{
		"key": n.NetmonkServerKey,
	})
	reqBody := bytes.NewBuffer(postBody)

	endpoint := fmt.Sprintf(verifyEndpoint, n.NetmonkServerID)
	resp, err := http.Post(fmt.Sprintf("%s%s", n.NetmonkHost, endpoint), "application/json", reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		cc := &CustomerCredentials{}
		err = ParseResponseFormat(resp, cc)
		if err != nil {
			return nil, err
		}

		return cc, nil
	}

	return nil, fmt.Errorf("failed to verify agent")
}

// ParseResponseFormat parse default response format netmonk API service
func ParseResponseFormat(r *http.Response, content interface{}) error {
	body := &ResponseFormat{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&body)
	if err != nil {
		return fmt.Errorf("encode body: %w", err)
	}

	encodedContent, _ := json.Marshal(body.Data)
	err = json.Unmarshal(encodedContent, &content)
	if err != nil {
		return fmt.Errorf("unmarshal content: %w", err)
	}

	return nil
}
