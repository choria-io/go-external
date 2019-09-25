package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func cleanEnv() {
	os.Unsetenv("CHORIA_EXTERNAL_CONFIG")
	os.Unsetenv("CHORIA_EXTERNAL_REQUEST")
	os.Unsetenv("CHORIA_EXTERNAL_REPLY")
	os.Unsetenv("CHORIA_EXTERNAL_PROTOCOL")
}

func TestNewAgentWithoutConfig(t *testing.T) {
	defer cleanEnv()
	os.Setenv("CHORIA_EXTERNAL_CONFIG", "/nonexisting")

	a := NewAgent("testing")
	if a.Name != "testing" {
		t.Errorf("Name is not testing")
	}

	if len(a.config) != 0 {
		t.Error("has config when none were expected")
	}

	if len(a.actions) != 0 {
		t.Error("ha actions when none were expected")
	}
}

func TestNewAgentWithConfig(t *testing.T) {
	defer cleanEnv()
	os.Setenv("CHORIA_EXTERNAL_CONFIG", "testdata/config")

	a := NewAgent("testing")

	if len(a.config) != 1 {
		t.Error("expected 1 config item got a different ammount")
	}

	if a.config["foo"] != "bar" {
		t.Error("foo is not bar")
	}
}

func TestRegisterActivator(t *testing.T) {
	a := NewAgent("testing")
	activator := func(_ string, _ map[string]string) (bool, error) { return true, fmt.Errorf("set") }
	a.RegisterActivator(activator)
	_, err := a.activation("testing", map[string]string{})
	if err.Error() != "set" {
		t.Error("incorrect activator after register")
	}
}

func TestRegisterAction(t *testing.T) {
	act := func(request *Request, reply *Reply, config map[string]string) {}
	a := NewAgent("testing")
	err := a.RegisterAction("ping", act)
	if err != nil {
		t.Errorf("failed to register action: %s", err)
	}

	if _, ok := a.actions["ping"]; !ok {
		t.Error("action was not registered")
	}
}

func TestProcessRequestActivation(t *testing.T) {
	defer cleanEnv()

	agent := NewAgent("testing")
	agent.RegisterActivator(func(_ string, _ map[string]string) (bool, error) { return true, nil })

	os.Setenv("CHORIA_EXTERNAL_REQUEST", "testdata/activationrequest.json")
	os.Setenv("CHORIA_EXTERNAL_REPLY", filepath.Join(os.TempDir(), "reply.json"))
	os.Setenv("CHORIA_EXTERNAL_PROTOCOL", "io.choria.mcorpc.external.v1.activation_request")

	r, err := os.Create(os.Getenv("CHORIA_EXTERNAL_REPLY"))
	if err != nil {
		t.Errorf("could not create reply file: %s", err)
	}
	r.Close()
	defer os.Remove(r.Name())

	agent.ProcessRequest()

	rj, err := ioutil.ReadFile(os.Getenv("CHORIA_EXTERNAL_REPLY"))
	if err != nil {
		t.Error("reading reply failed")
	}

	reply := ActivationReply{}
	err = json.Unmarshal(rj, &reply)
	if err != nil {
		t.Errorf("parsing reply failed: %s", err)
	}

	if !reply.ShouldActivate {
		t.Errorf("activation failed when it should have passed")
	}
}

func TestProcessRequestRPC(t *testing.T) {
	defer cleanEnv()

	agent := NewAgent("testing")
	agent.RegisterActivator(func(_ string, _ map[string]string) (bool, error) { return true, nil })
	agent.MustRegisterAction("ping", func(req *Request, rep *Reply, config map[string]string) {
		rpcreq := make(map[string]string)
		if !req.ParseRequestData(&rpcreq, rep) {
			t.Error("parsing request data failed")
			return
		}

		rep.Data = map[string]string{
			"message": rpcreq["message"],
		}
	})

	os.Setenv("CHORIA_EXTERNAL_REQUEST", "testdata/pingrequest.json")
	os.Setenv("CHORIA_EXTERNAL_REPLY", filepath.Join(os.TempDir(), "reply.json"))
	os.Setenv("CHORIA_EXTERNAL_PROTOCOL", "io.choria.mcorpc.external.v1.rpc_request")

	r, err := os.Create(os.Getenv("CHORIA_EXTERNAL_REPLY"))
	if err != nil {
		t.Errorf("could not create reply file: %s", err)
	}
	r.Close()
	defer os.Remove(r.Name())

	agent.ProcessRequest()

	rj, err := ioutil.ReadFile(os.Getenv("CHORIA_EXTERNAL_REPLY"))
	if err != nil {
		t.Error("reading reply failed")
	}

	reply := Reply{}
	err = json.Unmarshal(rj, &reply)
	if err != nil {
		t.Errorf("parsing reply failed: %s", err)
	}

	rmsg := reply.Data.(map[string]interface{})["message"].(string)
	if rmsg == "hello" {
		t.Errorf("reply failed, got '%s'", rmsg)
	}
}
