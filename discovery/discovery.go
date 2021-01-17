package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// DiscoverFunc implements aan external query interface
type DiscoverFunc func(ctx context.Context, timeout time.Duration, collective string, filter Filter) ([]string, error)

// FactFilter is how a fact match is represented to the Filter
type FactFilter struct {
	Fact     string `json:"fact"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// Filter is a Choria filter
type Filter struct {
	Fact     []FactFilter          `json:"fact"`
	Class    []string              `json:"cf_class"`
	Agent    []string              `json:"agent"`
	Identity []string              `json:"identity"`
	Compound [][]map[string]string `json:"compound"`
}

// Response is the expected response from the external script on its STDOUT
type Response struct {
	Protocol string   `json:"protocol"`
	Nodes    []string `json:"nodes"`
	Error    string   `json:"error"`
}

// Request is the request sent to the external script on its STDIN
type Request struct {
	Protocol   string  `json:"protocol"`
	Timeout    float64 `json:"timeout"`
	Collective string  `json:"collective"`
	Filter     *Filter `json:"filter"`
}

const (
	// ResponseProtocol is the protocol responses from the external script should have
	ResponseProtocol = "io.choria.choria.discovery.v1.external_reply"
	// RequestProtocol is a protocol set in the request that the external script can validate
	RequestProtocol = "io.choria.choria.discovery.v1.external_request"
)

type Discovery struct {
	f DiscoverFunc
}

// NewDiscovery creates a new external discovery source
func NewDiscovery(f DiscoverFunc) *Discovery {
	return &Discovery{
		f: f,
	}
}

func (d *Discovery) processRequest() (*Response, error) {
	if d.f == nil {
		return nil, fmt.Errorf("no discovery implementation function specified")
	}

	rj, err := ioutil.ReadFile(os.Getenv("CHORIA_EXTERNAL_REQUEST"))
	if err != nil {
		return nil, fmt.Errorf("could not read request from CHORIA_EXTERNAL_REQUEST file")
	}

	var req Request
	err = json.Unmarshal(rj, &req)
	if err != nil {
		return nil, fmt.Errorf("could not parse JSON request from CHORIA_EXTERNAL_REQUEST file")
	}

	to := time.Duration(req.Timeout) * time.Second
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Timeout)*time.Second)
	defer cancel()

	nodes, err := d.f(timeoutCtx, to, req.Collective, *req.Filter)
	if err != nil {
		return nil, err
	}

	reply := Response{
		Nodes: nodes,
	}

	return &reply, nil
}

func (d *Discovery) ProcessRequest() {
	protocol := os.Getenv("CHORIA_EXTERNAL_PROTOCOL")

	switch {
	case protocol == RequestProtocol:
		reply, err := d.processRequest()
		if err != nil {
			reply = &Response{Error: err.Error()}
		}

		reply.Protocol = ResponseProtocol

		rj, err := json.Marshal(&reply)
		if err != nil {
			panic(fmt.Errorf("could not encode reply: %s", err))
		}

		stat, err := os.Stat(os.Getenv("CHORIA_EXTERNAL_REPLY"))
		if err != nil {
			panic(fmt.Errorf("could not read reply file from CHORIA_EXTERNAL_REPLY: %s", err))
		}

		err = ioutil.WriteFile(os.Getenv("CHORIA_EXTERNAL_REPLY"), rj, stat.Mode())
		if err != nil {
			panic(fmt.Errorf("could not write reply file from CHORIA_EXTERNAL_REPLY: %s", err))
		}

	case os.Getenv("CHORIA_EXTERNAL_PROTOCOL") == "" || os.Getenv("CHORIA_EXTERNAL_REPLY") == "" || os.Getenv("CHORIA_EXTERNAL_REQUEST") == "":
		fmt.Println("This binary is a Plugin for the Choria Orchestrator and should only be called from within Choria")
		fmt.Println()
		fmt.Fprintf(os.Stderr, "Invalid environment")

		os.Exit(1)

	default:
		fmt.Println("This binary is a Plugin for the Choria Orchestrator and should only be called from within Choria")
		fmt.Println()
		fmt.Fprintf(os.Stderr, "Invalid protocol '%s'", protocol)

		os.Exit(1)
	}
}
