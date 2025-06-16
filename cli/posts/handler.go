package posts

import (
	"fmt"
	"github.com/oullin/pkg/markdown"
)

func (h *Handler) HandlePost(post *markdown.Post) {
	fmt.Println("HandlePost!")
}
