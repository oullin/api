package menu

import (
	"fmt"
	"github.com/oullin/cli/posts"
	"github.com/oullin/pkg/cli"
	"golang.org/x/term"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func (p *Panel) PrintLine() {
	_, _ = p.Reader.ReadString('\n')
}

func (p *Panel) GetChoice() int {
	if p.Choice == nil {
		return 0
	}

	return *p.Choice
}

func (p *Panel) CaptureInput() error {
	fmt.Print(cli.YellowColour + "Select an option: " + cli.Reset)
	input, err := p.Reader.ReadString('\n')

	if err != nil {
		return fmt.Errorf("%s error reading input: %v %s", cli.RedColour, err, cli.Reset)
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)

	if err != nil {
		return fmt.Errorf("%s Please enter a valid number. %s", cli.RedColour, cli.Reset)
	}

	p.Choice = &choice

	return nil
}

func (p *Panel) PrintMenu() {
	// Try to get the terminal width; default to 80 if it fails
	width, _, err := term.GetSize(int(os.Stdout.Fd()))

	if err != nil || width < 20 {
		width = 80
	}

	inner := width - 2 // space between the two border chars

	// Build box pieces
	border := "╔" + strings.Repeat("═", inner) + "╗"
	title := "║" + p.CenterText(" Main Menu ", inner) + "║"
	divider := "╠" + strings.Repeat("═", inner) + "╣"
	footer := "╚" + strings.Repeat("═", inner) + "╝"

	// Print in color
	fmt.Println()
	fmt.Println(cli.CyanColour + border)
	fmt.Println(title)
	fmt.Println(divider)

	p.PrintOption("1) Parse Posts", inner)
	p.PrintOption("2) Show Time", inner)
	p.PrintOption("3) Show Date", inner)
	p.PrintOption("0) Exit", inner)

	fmt.Println(footer + cli.Reset)
}

// PrintOption left-pads a space, writes the text, then fills to the full inner width.
func (p *Panel) PrintOption(text string, inner int) {
	content := " " + text
	if len(content) > inner {
		content = content[:inner]
	}
	padding := inner - len(content)
	fmt.Printf("║%s%s║\n", content, strings.Repeat(" ", padding))
}

// CenterText centers s within width, padding with spaces.
func (p *Panel) CenterText(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}

	pad := width - len(s)
	left := pad / 2
	right := pad - left

	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func (p *Panel) CapturePostURL() (*posts.Input, error) {
	fmt.Print("Enter the post markdown file URL: ")
	uri, err := p.Reader.ReadString('\n')

	if err != nil {
		return nil, fmt.Errorf("%sError reading the given post URL: %v %s", cli.RedColour, err, cli.Reset)
	}

	uri = strings.TrimSpace(uri)
	if uri == "" {
		return nil, fmt.Errorf("%sError: no URL provided: %s", cli.RedColour, cli.Reset)
	}

	parsedURL, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("%sError: invalid URL: %v %s", cli.RedColour, err, cli.Reset)
	}

	if parsedURL.Scheme != "https" || parsedURL.Host != "raw.githubusercontent.com" {
		return nil, fmt.Errorf("%sError: URL must begin with https://raw.githubusercontent.com %s", cli.RedColour, cli.Reset)
	}

	input := posts.Input{Url: parsedURL.String()}
	validate := p.Validator

	if _, err := validate.Rejects(input); err != nil {
		return nil, fmt.Errorf(
			"%sError validating the given post URL: %v %s \n%sViolations:%s %s",
			cli.RedColour,
			err,
			cli.Reset,
			cli.BlueColour,
			cli.Reset,
			validate.GetErrorsAsJason(),
		)
	}

	return &input, nil
}
