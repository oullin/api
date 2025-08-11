package posts

import (
	"context"
	"fmt"
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/pkg/markdown"
	"github.com/oullin/pkg/portal"
	"net/http"
	"time"
)

type Handler struct {
	Input       *Input
	Client      *portal.Client
	Posts       *repository.Posts
	Users       *repository.Users
	IsDebugging bool
}

func MakeHandler(input *Input, client *portal.Client, db *database.Connection) Handler {
	tags := &repository.Tags{DB: db}
	categories := &repository.Categories{DB: db}

	return Handler{
		Input:       input,
		IsDebugging: false,
		Client:      client,
		Users:       &repository.Users{DB: db},
		Posts:       &repository.Posts{DB: db, Categories: categories, Tags: tags},
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
	fmt.Printf("Categories: %s\n", post.Categories)
	fmt.Printf("Tags Alt: %s\n", post.Tags)
	fmt.Println("\n--- Content ---")
	fmt.Println(post.Content)
}
