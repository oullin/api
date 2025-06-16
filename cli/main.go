package main

import (
	"bufio"
	"fmt"
	"github.com/oullin/cli/menu"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/cli"
	"os"
	"time"
)

func main() {
	panel := menu.Panel{
		Reader:    bufio.NewReader(os.Stdin),
		Validator: pkg.GetDefaultValidator(),
	}

	panel.PrintMenu()

	for {
		err := panel.CaptureInput()

		if err != nil {
			cli.MakeTextColour(err.Error(), cli.Red).Println()
			continue
		}

		switch panel.GetChoice() {
		case 1:
			input, err := panel.CapturePostURL()

			if err != nil {
				fmt.Println(err)
				continue
			}

			err = input.Parse()

			if err != nil {
				fmt.Println(err)
				continue
			}

			return
		case 2:
			showTime()
		case 0:
			fmt.Println(cli.Green + "Goodbye!" + cli.Reset)
			return
		default:
			fmt.Println(cli.Red, "Unknown option. Try again.", cli.Reset)
		}

		fmt.Print("\nPress Enter to continue...")

		panel.PrintLine()
	}
}

func showTime() {
	now := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(cli.Green, "\nCurrent time is", now, cli.Reset)
}
