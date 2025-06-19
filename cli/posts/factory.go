package posts

import (
	"context"
	"fmt"
	"github.com/oullin/boost"
	"github.com/oullin/database/repository"
	"github.com/oullin/env"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/markdown"
	"net/http"
	"time"
)

type Handler struct {
	Input       *Input
	Client      *pkg.Client
	Posts       *repository.Posts
	Users       *repository.Users
	IsDebugging bool
}

func MakeHandler(input *Input, client *pkg.Client, env *env.Environment) Handler {
	db := boost.MakeDbConnection(env)

	return Handler{
		Input:       input,
		Client:      client,
		IsDebugging: false,
		Posts: &repository.Posts{
			DB: db,
			Categories: &repository.Categories{
				DB: db,
			},
		},
		Users: &repository.Users{
			DB: db,
		},
	}
}

func (h Handler) NotParsed() (bool, error) {
	var err error
	var content string
	uri := h.Input.Url

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	h.Client.OnHeaders = func(req *http.Request) {
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Pragma", "no-cache")
	}

	content, err = h.Client.Get(ctx, uri)

	if err != nil {
		return false, fmt.Errorf("error fetching url [%s]: %w", uri, err)
	}

	var article *markdown.Post
	if article, err = markdown.Parse(content); err != nil || article == nil {
		return false, fmt.Errorf("error parsing url [%s]: %w", uri, err)
	}

	if h.IsDebugging {
		h.RenderArticle(article)
	}

	if err = h.HandlePost(article); err != nil {
		return true, err
	}

	return true, nil
}

func (h Handler) RenderArticle(post *markdown.Post) {
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
}
