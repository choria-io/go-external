package agent

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type externalAgent struct{}

func (e externalAgent) publishReply(rep interface{}) error {
	repfile := os.Getenv("CHORIA_EXTERNAL_REPLY")
	if !fileExist(repfile) {
		return fmt.Errorf("request file '%s' from CHORIA_EXTERNAL_REPLY does not exist", repfile)
	}

	j, err := json.Marshal(rep)
	if err != nil {
		return fmt.Errorf("could not JSON encode reply data: %s", err)
	}

	f, err := os.Create(repfile)
	if err != nil {
		return fmt.Errorf("could not open reply file %s: %s", repfile, err)
	}
	defer f.Close()

	_, err = fmt.Fprint(f, string(j))
	if err != nil {
		return fmt.Errorf("failed writing to reply file %s: %s", repfile, err)
	}

	return nil
}

func (e externalAgent) loadRequest(protocol string, req interface{}) error {
	reqproto := os.Getenv("CHORIA_EXTERNAL_PROTOCOL")
	reqfile := os.Getenv("CHORIA_EXTERNAL_REQUEST")

	if os.Getenv("CHORIA_EXTERNAL_PROTOCOL") != protocol {
		return fmt.Errorf("unexpected protocol '%s'", reqproto)
	}

	if !fileExist(reqfile) {
		return fmt.Errorf("request file '%s' from CHORIA_EXTERNAL_REQUEST does not exist", reqfile)
	}

	reqj, err := ioutil.ReadFile(reqfile)
	if err != nil {
		return fmt.Errorf("could not load request: %s", err)
	}

	err = json.Unmarshal(reqj, req)
	if err != nil {
		return fmt.Errorf("could ot parse request: %s", err)
	}

	return nil
}
