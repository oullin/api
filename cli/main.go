package main

import (
	"bufio"
	"fmt"
	"github.com/oullin/boost"
	"github.com/oullin/cli/menu"
	"github.com/oullin/cli/posts"
	"github.com/oullin/env"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/cli"
	"os"
	"time"
)

var environment *env.Environment

func init() {
	secrets, _ := boost.Spark("./../.env")

	environment = secrets
}

func main() {
	postsHandler := posts.MakePostsHandler(environment)

	panel := menu.Panel{
		Reader:    bufio.NewReader(os.Stdin),
		Validator: pkg.GetDefaultValidator(),
	}

	panel.PrintMenu()

	for {
		err := panel.CaptureInput()

		if err != nil {
			fmt.Println(cli.Red + err.Error() + cli.Reset)
			continue
		}

		switch panel.GetChoice() {
		case 1:
			input, err := panel.CapturePostURL()

			if err != nil {
				fmt.Println(err)
				continue
			}

			if post, err := input.Parse(); err != nil {
				fmt.Println(err)
				continue
			} else {
				(*postsHandler).HandlePost(post)
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
	fmt.Println("")
	now := time.Now().Format("2006-01-02 15:04:05")
	fmt.Println(cli.Green, "\nCurrent time is", now, cli.Reset)
}
