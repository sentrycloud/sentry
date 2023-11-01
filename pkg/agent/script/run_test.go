package script

import (
	"testing"
)

func TestRunCommand(t *testing.T) {
	out, err := RunCommand(3, "python3", "test.py")
	if err != nil {
		t.Errorf("run command failed: %v\n", err)
	}

	if string(out)[0:5] != "hello" {
		t.Errorf("not expect output\n")
	}

	out, err = RunCommand(3, "python3", "test.py", "block")
	if err == nil {
		t.Errorf("not expect result")
	}
	t.Logf("run command error: %v\n", err)

	out, err = RunCommand(3, "python3", "notExist.py")
	if err == nil {
		t.Errorf("not expect result")
	}
	t.Logf("run command error: %v\n", err)
}
