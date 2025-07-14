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

	categories := h.ParseCategories(payload)
	if len(categories) == 0 {
		return fmt.Errorf("handler: the given categories [%v] are empty", payload.Categories)
	}

	attrs := database.PostsAttrs{
		AuthorID:    author.ID,
		PublishedAt: publishedAt,
		Slug:        payload.Slug,
		Title:       payload.Title,
		Excerpt:     payload.Excerpt,
		Content:     payload.Content,
		ImageURL:    payload.ImageURL,
		Categories:  categories,
		Tags:        h.ParseTags(payload),
	}

	if _, err = h.Posts.Create(attrs); err != nil {
		return fmt.Errorf("handler: error persiting the post [%s]: %s", attrs.Title, err.Error())
	}

	cli.Successln("\n" + fmt.Sprintf("Post [%s] created successfully.", attrs.Title))

	return nil
}

func (h Handler) ParseCategories(payload *markdown.Post) []database.CategoriesAttrs {
	var categories []database.CategoriesAttrs
	parts := strings.Split(payload.Categories, ",")

	for _, category := range parts {
		slug := strings.TrimSpace(strings.ToLower(category))

		if item := h.Posts.FindCategoryBy(slug); item != nil {
			categories = append(categories, database.CategoriesAttrs{
				Slug:        item.Slug,
				Name:        item.Name,
				Id:          item.ID,
				Description: item.Description,
			})
		}
	}

	return categories
}

func (h Handler) ParseTags(payload *markdown.Post) []database.TagAttrs {
	var tags []database.TagAttrs

	for _, tag := range payload.Tags {
		slug := strings.TrimSpace(strings.ToLower(tag))

		if item := h.Posts.FindTagBy(slug); item != nil {
			tags = append(tags, database.TagAttrs{
				Id:   item.ID,
				Slug: slug,
				Name: slug,
			})
		}
	}

	return tags
}
