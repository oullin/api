package projects

import (
	"context"
)

type PublishedAtResolver func(context.Context, ProjectsData) (string, error)
