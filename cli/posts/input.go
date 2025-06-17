package posts

import (
	"fmt"
	"github.com/oullin/pkg/cli"
	"github.com/oullin/pkg/markdown"
)

type Input struct {
	Url string `validate:"required,min=10"`
}

func (i *Input) Parse() (*markdown.Post, error) {
	file := markdown.Parser{
		Url: i.Url,
	}

	response, err := file.Fetch()

	if err != nil {
		return nil, fmt.Errorf("%sError fetching the markdown content: %v %s", cli.Red, err, cli.Reset)
	}

	post, err := markdown.Parse(response)

	if err != nil {
		return nil, fmt.Errorf("%sEerror parsing markdown: %v %s", cli.Red, err, cli.Reset)
	}

	// --- All good!
	// Todo: Save post in the DB.
	fmt.Printf("Title: %s\n", post.Title)
	fmt.Printf("Excerpt: %s\n", post.Excerpt)
	fmt.Printf("Slug: %s\n", post.Slug)
	fmt.Printf("Author: %s\n", post.Author)
	fmt.Printf("Image URL: %s\n", post.ImageURL)
	fmt.Printf("Image Alt: %s\n", post.ImageAlt)
	fmt.Printf("Category: %s\n", post.Category)
	fmt.Printf("Category Slug: %s\n", post.CategorySlug)
	fmt.Printf("Tags Alt: %s\n", post.Tags)
	fmt.Println("\n--- Content ---")
	fmt.Println(post.Content)

	return &post, nil
}
