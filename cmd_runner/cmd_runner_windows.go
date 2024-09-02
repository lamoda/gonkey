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
	// by default timeout should be 3s
	if timeout <= 0 {
		timeout = 3
	}
	cmd := exec.Command(strings.TrimRight(scriptPath, "\n"))
	cmd.Env = os.Environ()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(time.Duration(timeout) * time.Second):
		// Kill process
		if err := cmd.Process.Kill(); err != nil {
			return err
		}
		fmt.Printf("Process killed as timeout(%d) reached\n", timeout)
	case err := <-done:
		if err != nil {
			return fmt.Errorf("process finished with error = %v", err)
		}
		log.Print("Process finished successfully")
	}

	// Print log
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		m := scanner.Text()
		fmt.Println(m)
	}

	return nil
}
