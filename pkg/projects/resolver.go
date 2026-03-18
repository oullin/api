package projects

import (
	"context"

	"github.com/oullin/handler/payload"
)

type PublishedAtResolver func(context.Context, payload.ProjectsData) (string, error)
