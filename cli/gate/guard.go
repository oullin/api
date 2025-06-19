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
	salt   *string
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
		return fmt.Errorf("error reading input: %v", err)
	}

	input = strings.TrimSpace(input)

	if len(input) == 0 {
		return fmt.Errorf("token cannot be empty")
	}

	if len(input) > 1024 {
		return fmt.Errorf("token is too long")
	}

	g.salt = &input

	return nil
}

func (g *Guard) Rejects() bool {
	if g.salt == nil {
		return true
	}

	salt := *g.salt

	return g.token.IsInvalid(salt)
}
