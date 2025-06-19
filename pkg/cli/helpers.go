package cli

import (
	"fmt"
	"os"
	"os/exec"
)

// ClearScreen attempts to clear the terminal screen by running the "clear" command.
// If the command fails, it logs an error message.
func ClearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		message := fmt.Sprintf("Could not clear screen. Error: %s", err.Error())

		Errorln(message)
	}
}
