package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"
)

func cleanEnv() {
	os.Unsetenv("CHORIA_EXTERNAL_REQUEST")
	os.Unsetenv("CHORIA_EXTERNAL_REPLY")
	os.Unsetenv("CHORIA_EXTERNAL_PROTOCOL")
}

func setupExecution(t *testing.T, req Request) (*os.File, *os.File) {
	t.Helper()

	reqfile, err := ioutil.TempFile("", "request")
	if err != nil {
		t.Fatalf("could not create request temp file: %s", err)
	}
	rj, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %s", err)
	}

	_, err = reqfile.Write(rj)
	if err != nil {
		t.Fatalf("write failed: %s", err)
	}
	reqfile.Close()

	repfile, err := ioutil.TempFile("", "reply")
	if err != nil {
		t.Fatalf("could not create reply temp file: %s", err)
	}
	repfile.Close()

	os.Setenv("CHORIA_EXTERNAL_PROTOCOL", RequestProtocol)
	os.Setenv("CHORIA_EXTERNAL_REPLY", repfile.Name())
	os.Setenv("CHORIA_EXTERNAL_REQUEST", reqfile.Name())

	return reqfile, repfile
}

func readResponse(t *testing.T, f string) *Response {
	t.Helper()

	rj, err := ioutil.ReadFile(f)
	if err != nil {
		t.Fatalf("could not read reply: %s", err)
	}

	var reply Response
	err = json.Unmarshal(rj, &reply)
	if err != nil {
		t.Fatalf("could not parse reply: %s", err)
	}

	return &reply
}

func newRequest() Request {
	return Request{
		Protocol:   RequestProtocol,
		Timeout:    2,
		Collective: "mcollective",
		Filter: &Filter{
			Fact:     []FactFilter{},
			Class:    []string{"development"},
			Agent:    []string{},
			Identity: []string{},
			Compound: [][]map[string]string{},
		},
	}
}

func TestDiscover(t *testing.T) {
	cleanEnv()

	d := NewDiscovery(func(ctx context.Context, timeout time.Duration, collective string, filter Filter) ([]string, error) {
		return nil, fmt.Errorf("not implemented")
	})

	reqfile, repfile := setupExecution(t, newRequest())
	defer os.Remove(reqfile.Name())
	defer os.Remove(repfile.Name())

	d.ProcessRequest()
	reply := readResponse(t, repfile.Name())

	if reply.Error != "not implemented" {
		t.Fatalf("expected error got none")
	}

	d = NewDiscovery(func(ctx context.Context, timeout time.Duration, collective string, filter Filter) ([]string, error) {
		if collective != "mcollective" {
			return nil, fmt.Errorf("invalid collective")
		}

		if filter.Class[0] != "development" {
			return nil, fmt.Errorf("invalid filter")
		}

		return []string{"one", "two", "three"}, nil
	})

	d.ProcessRequest()
	reply = readResponse(t, repfile.Name())
	if reply.Protocol != ResponseProtocol {
		t.Fatalf("incorrect reponse protocol, wanted %q, got %q", ResponseProtocol, reply.Protocol)
	}

	if reply.Error != "" {
		t.Fatal(reply.Error)
	}

	if !reflect.DeepEqual(reply.Nodes, []string{"one", "two", "three"}) {
		t.Fatalf("incorrect nodes received")
	}
}
