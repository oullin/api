package gate

import (
	"bufio"
	"fmt"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/cli"
	"os"
	"strings"
)

type Guard struct {
	salt   string
	token  auth.Token
	reader *bufio.Reader
}

func MakeGuard(token auth.Token) Guard {
	return Guard{
		token:  token,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (g *Guard) CaptureInput() error {
	cli.Warning("Type the public token: ")
	input, err := g.reader.ReadString('\n')

	if err != nil {
		return fmt.Errorf("%s error reading input: %v %s", cli.RedColour, err, cli.Reset)
	}

	g.salt = strings.TrimSpace(input)

	return nil
}

func (g *Guard) Rejects() bool {
	return g.token.IsInvalid(g.salt)
}
