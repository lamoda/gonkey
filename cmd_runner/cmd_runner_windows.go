package cmd_runner

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func CmdRun(scriptPath string, timeout int) error {
	output, err := CmdRunWithIO(scriptPath, "", timeout)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

func CmdRunWithIO(scriptPath string, input string, timeout int) (string, error) {
	//by default timeout should be 3s
	if timeout <= 0 {
		timeout = 3
	}

	args := strings.Split(strings.TrimRight(scriptPath, "\n"), " ")
	script := args[0]
	if len(args) > 1 {
		args = args[1:]
	} else {
		args = args[:0]
	}

	var output string

	cmd := exec.Command(script, args...)
	cmd.Env = os.Environ()
	if len(input) > 0 {
		cmd.Stdin = strings.NewReader(input)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return output, err
	}

	if err := cmd.Start(); err != nil {
		return output, err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			output += scanner.Text()
		}
	}()

	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		// Kill process
		if err := cmd.Process.Kill(); err != nil {
			return output, err
		}
		log.Printf("Process killed as timeout(%d) reached\n", timeout)
	case err := <-done:
		if err != nil {
			return output, fmt.Errorf("process finished with error = %v", err)
		}
		log.Print("Process finished successfully")
	}

	return output, nil
}
