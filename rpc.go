package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	rpcRequestProtocol = "io.choria.mcorpc.external.v1.rpc_request"
	rpcReplyProtocol   = "io.choria.mcorpc.external.v1.rpc_reply"
)

// ActionHandler is a function that implements a RPC action
type ActionHandler func(req *Request, rep *Reply, config map[string]string)

type rpc struct {
	externalAgent
	actions map[string]ActionHandler
	config  map[string]string
}

func newRPC(actions map[string]ActionHandler, config map[string]string) (*rpc, error) {
	return &rpc{actions: actions, config: config}, nil
}

func (r *rpc) panicIfError(err error, format string, a ...interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, format, a...)
		os.Exit(1)
	}
}

func (r *rpc) fail(format string, a ...interface{}) bool {
	reply := &Reply{
		Statuscode: Aborted,
		Statusmsg:  fmt.Sprintf(format, a...),
		Data:       make(map[string]interface{}),
	}

	err := r.publishReply(reply)
	r.panicIfError(err, "could not write reply: %s", err)

	return true
}

func (r *rpc) failIfError(err error, format string, a ...interface{}) bool {
	if err == nil {
		return false
	}

	return r.fail(format, a...)
}

func (r *rpc) hasAction(action string) bool {
	_, ok := r.actions[action]
	return ok
}

func (r *rpc) handleRequest() error {
	request := &Request{}
	reply := &Reply{}

	jreq, err := ioutil.ReadFile(os.Getenv("CHORIA_EXTERNAL_REQUEST"))
	if r.failIfError(err, "could not read request from CHORIA_EXTERNAL_REQUEST file") {
		return nil
	}

	if r.failIfError(json.Unmarshal(jreq, request), "could not parse request") {
		return nil
	}

	if request.Action == "" {
		if r.failIfError(fmt.Errorf("invalid action"), "request failed") {
			return nil
		}
	}

	action, ok := r.actions[request.Action]
	if action == nil || !ok {
		if r.fail("unknown action %s", request.Action) {
			return nil
		}
	}

	action(request, reply, r.config)

	err = r.publishReply(reply)
	r.panicIfError(err, "request failed: %s", err)

	return nil
}
