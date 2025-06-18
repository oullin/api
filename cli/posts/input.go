package posts

import (
	"fmt"
	"github.com/oullin/pkg/cli"
	"github.com/oullin/pkg/markdown"
)

type Input struct {
	Url          string `validate:"required,min=10"`
	Debug        bool
	MarkdownPost *markdown.Post
}

func (i *Input) Parse() (*markdown.Post, error) {
	file := markdown.Parser{
		Url: i.Url,
	}

	response, err := file.Fetch()

	if err != nil {
		return nil, fmt.Errorf("%sError fetching the markdown content: %v %s", cli.RedColour, err, cli.Reset)
	}

	post, err := markdown.Parse(response)
	i.MarkdownPost = &post

	if err != nil {
		return nil, fmt.Errorf("%sEerror parsing markdown: %v %s", cli.RedColour, err, cli.Reset)
	}

	return &post, nil
}

func (i *Input) Render() {
	fmt.Printf("Title: %s\n", i.MarkdownPost.Title)
	fmt.Printf("Excerpt: %s\n", i.MarkdownPost.Excerpt)
	fmt.Printf("Slug: %s\n", i.MarkdownPost.Slug)
	fmt.Printf("Author: %s\n", i.MarkdownPost.Author)
	fmt.Printf("Image URL: %s\n", i.MarkdownPost.ImageURL)
	fmt.Printf("Image Alt: %s\n", i.MarkdownPost.ImageAlt)
	fmt.Printf("Category: %s\n", i.MarkdownPost.Category)
	fmt.Printf("Category Slug: %s\n", i.MarkdownPost.CategorySlug)
	fmt.Printf("Tags Alt: %s\n", i.MarkdownPost.Tags)
	fmt.Println("\n--- Content ---")
	fmt.Println(i.MarkdownPost.Content)
}
