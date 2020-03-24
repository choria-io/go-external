package agent

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// Agent is a Choria External agent helper library that assist you with building
// agents in Go that does not need to be compiled into the Choria binary
type Agent struct {
	Name       string
	activation ActivationHandler
	actions    map[string]ActionHandler
	config     map[string]string
}

// NewAgent creates a new agent
func NewAgent(name string) *Agent {
	a := &Agent{
		Name:    name,
		config:  make(map[string]string),
		actions: make(map[string]ActionHandler),
	}

	err := a.parseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse configuration: %s", err)
		os.Exit(1)
	}

	return a
}

// FactsPath returns the path to the node facts, empty string when not provided
func FactsPath() string {
	return os.Getenv("CHORIA_EXTERNAL_FACTS")
}

// Facts returns the server facts provided during invocation, empty JSON hash when not provided
func Facts() (json.RawMessage, error) {
	ff := os.Getenv("CHORIA_EXTERNAL_FACTS")
	if ff == "" {
		return []byte(`{}`), nil
	}

	fj, err := ioutil.ReadFile(ff)
	if err != nil {
		return []byte(`{}`), err
	}

	return fj, nil
}

// RegisterActivator registers a function used to check if the agent should be active,
// with no activator set the agent will always activate
func (a *Agent) RegisterActivator(handler ActivationHandler) {
	a.activation = handler
}

// RegisterAction registers a new action
func (a *Agent) RegisterAction(action string, handler ActionHandler) error {
	_, ok := a.actions[action]
	if ok {
		return fmt.Errorf("duplicate action %s", action)
	}

	a.actions[action] = handler

	return nil
}

// MustRegisterAction registers an action and panics if any error occur
func (a *Agent) MustRegisterAction(action string, handler ActionHandler) {
	err := a.RegisterAction(action, handler)
	if err != nil {
		panic(err)
	}
}

// ProcessRequest processes an incoming request
func (a *Agent) ProcessRequest() {
	protocol := os.Getenv("CHORIA_EXTERNAL_PROTOCOL")

	switch protocol {
	case activationProtocol:
		a.processActivation()

	case rpcRequestProtocol:
		a.processRPC()

	default:
		fmt.Println("This binary is a Plugin for the Choria Orchestrator and should only be called from within Choria")
		fmt.Println()
		fmt.Fprintf(os.Stderr, "Invalid protocol '%s'", protocol)

		os.Exit(1)
	}
}

func (a *Agent) parseConfig() error {
	configpath := os.Getenv("CHORIA_EXTERNAL_CONFIG")
	if configpath == "" || !fileExist(configpath) {
		return nil
	}

	file, err := os.Open(configpath)
	if err != nil {
		return err
	}
	defer file.Close()

	itemr := regexp.MustCompile(`(.+?)\s*=\s*(.+)`)
	skipr := regexp.MustCompile(`^#|^$`)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if skipr.MatchString(line) || !itemr.MatchString(line) {
			continue
		}

		matches := itemr.FindStringSubmatch(line)
		a.config[matches[1]] = matches[2]
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	return nil
}

func (a *Agent) defaultActivator(_ string, _ map[string]string) (bool, error) {
	return true, nil
}

func (a *Agent) processRPC() {
	rpch, err := newRPC(a.actions, a.config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create RPC handler: %s", err)
		os.Exit(1)
	}

	err = rpch.handleRequest()
	if err != nil {
		fmt.Fprintf(os.Stderr, "action failed: %s", err)
		os.Exit(1)
	}
}

func (a *Agent) processActivation() {
	if a.activation == nil {
		a.activation = a.defaultActivator
	}

	activator, err := newActivation(a.activation, a.config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create activation: %s", err)
		os.Exit(1)
	}

	err = activator.HandleRequest()
	if err != nil {
		fmt.Fprintf(os.Stderr, "activation failed: %s", err)
		os.Exit(1)
	}
}
