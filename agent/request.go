package agent

import (
	"encoding/json"
)

// Request is the request being published to the shim runner
type Request struct {
	Schema     string          `json:"$schema"`
	Protocol   string          `json:"protocol"`
	Agent      string          `json:"agent"`
	Action     string          `json:"action"`
	RequestID  string          `json:"requestid"`
	SenderID   string          `json:"senderid"`
	CallerID   string          `json:"callerid"`
	Collective string          `json:"collective"`
	TTL        int             `json:"ttl"`
	Time       int64           `json:"msgtime"`
	Data       json.RawMessage `json:"data"`
}

// ParseRequestData parses the RPC request JSON into target, sets reply to an appropriate failure code on error
func (r *Request) ParseRequestData(target interface{}, reply *Reply) bool {
	err := json.Unmarshal(r.Data, target)
	if err != nil {
		reply.InvalidData("Could not parse request data for %s#%s: %s", r.Agent, r.Action, err)
		return false
	}

	return true
}
