package gate

import (
	"bytes"
	"os/exec"
)

func RunCmd(cmdStr string, args ...string) error {
	if err := exec.Command(cmdStr, args...).Run(); err != nil {
		return err
	}
	return nil
}

func RunCmdOutput(cmdStr string, args ...string) string {
	cmd := exec.Command(cmdStr, args...)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stderr.String()
	}
	return out.String()

}
