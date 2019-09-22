package agent

import (
	"fmt"
	"os"
)

const (
	activationProtocol      = "io.choria.mcorpc.external.v1.activation_request"
	activationReplyProtocol = "io.choria.mcorpc.external.v1.activation_reply"
)

// ActivationHandler is user supplied logic to handle activation checks
type ActivationHandler func(agent string, config map[string]string) (bool, error)

// ActivationCheck is the request to determine if an agent should activate
type ActivationCheck struct {
	Schema   string `json:"$schema"`
	Protocol string `json:"protocol"`
	Agent    string `json:"agent"`

	handler ActivationHandler
	config  map[string]string

	externalAgent
}

// ActivationReply is the reply from the activation check message
type ActivationReply struct {
	ShouldActivate bool `json:"activate"`
}

func newActivation(h interface{}, config map[string]string) (*ActivationCheck, error) {
	handler, ok := h.(ActivationHandler)
	if !ok {
		return nil, fmt.Errorf("handler is not an ActivationHandler")
	}

	return &ActivationCheck{handler: handler, config: config}, nil
}

// HandleRequest handles the activation check
func (ac *ActivationCheck) HandleRequest() error {
	err := ac.loadRequest(activationProtocol, ac)
	if err != nil {
		Errorf("loading request failed: %s", err)
		os.Exit(1)
	}

	reply := &ActivationReply{}

	reply.ShouldActivate, err = ac.handler(ac.Agent, ac.config)
	if err != nil {
		Errorf("activation handler failed: %s", err)
		os.Exit(1)
	}

	err = ac.publishReply(reply)
	if err != nil {
		Errorf("publishing activation reply failed: %s", err)
		os.Exit(1)
	}

	return nil
}
