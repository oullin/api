package cli

import (
	"fmt"
	"os"
	"os/exec"
)

func ClearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		message := fmt.Sprintf("Could not clear screen. Error: %s", err.Error())

		Errorln(message)
	}
}
