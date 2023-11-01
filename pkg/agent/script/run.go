package script

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"syscall"
	"time"
)

func RunCommand(timeout int, scriptType string, args ...string) (out []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, scriptType, args...)
	var bOut bytes.Buffer
	var bError bytes.Buffer
	cmd.Stdout = &bOut
	cmd.Stderr = &bError
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err = cmd.Run()
	out = bOut.Bytes()
	if ctx.Err() == context.DeadlineExceeded {
		return out, errors.New("command timed out")
	}

	if err != nil {
		return out, err
	}

	if bError.Len() != 0 {
		err = errors.New(bError.String())
	}
	return out, err
}
