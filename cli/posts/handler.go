package posts

import (
	"fmt"
	"github.com/oullin/database"
	"github.com/oullin/pkg/cli"
	"github.com/oullin/pkg/markdown"
	"strings"
	"time"
)

func (h Handler) HandlePost(payload *markdown.Post) error {
	var err error
	var publishedAt *time.Time

	author := h.Users.FindBy(
		payload.Author,
	)

	if author == nil {
		return fmt.Errorf("handler: the given author [%s] does not exist", payload.Author)
	}

	if publishedAt, err = payload.GetPublishedAt(); err != nil {
		return fmt.Errorf("handler: the given published_at [%s] date is invalid", payload.PublishedAt)
	}

	attrs := database.PostsAttrs{
		AuthorID:    author.ID,
		PublishedAt: publishedAt,
		Slug:        payload.Slug,
		Title:       payload.Title,
		Excerpt:     payload.Excerpt,
		Content:     payload.Content,
		ImageURL:    payload.ImageURL,
		Categories:  h.ParseCategory(payload),
		Tags:        h.ParseTags(payload),
	}

	fmt.Printf("attrs: %v+n\n", attrs.Categories)

	panic("here ....")

	if _, err = h.Posts.Create(attrs); err != nil {
		return fmt.Errorf("handler: error persiting the post [%s]: %s", attrs.Title, err.Error())
	}

	cli.Successln("\n" + fmt.Sprintf("Post [%s] created successfully.", attrs.Title))

	return nil
}

// ParseCategory: Category is given like so (leadership:)
func (h Handler) ParseCategory(payload *markdown.Post) []database.CategoriesAttrs {
	var categories []database.CategoriesAttrs

	parts := strings.Split(payload.Category, ":")

	slice := append(categories, database.CategoriesAttrs{
		Slug: strings.Trim(parts[0], " "),
		Name: strings.Trim(parts[1], " "),
	})

	return slice
}

func (h Handler) ParseTags(payload *markdown.Post) []database.TagAttrs {
	var slice []database.TagAttrs

	for _, tag := range payload.Tags {
		slice = append(slice, database.TagAttrs{
			Slug: tag,
			Name: tag,
		})
	}

	return slice
}
