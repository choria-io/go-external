package agent

import (
	"testing"
)

func TestFileExist(t *testing.T) {
	if !fileExist("testdata/config") {
		t.Error("fileExist failed to find an existing file")
	}

	if fileExist("testdata/nonexisting") {
		t.Error("fileExist found a file that does not exist")
	}
}
