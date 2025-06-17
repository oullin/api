package posts

import (
    "fmt"
    "github.com/oullin/database"
    "github.com/oullin/pkg/markdown"
    "time"
)

func (h *Handler) HandlePost(payload *markdown.Post) error {
    var err error
    var publishedAt *time.Time
    author := h.UsersRepository.FindBy(payload.Author)

    if author == nil {
        return fmt.Errorf("the given author [%s] does not exist", payload.Author)
    }

    if publishedAt, err = payload.GetPublishedAt(); err != nil {
        return fmt.Errorf("the given published_at [%s] date is invalid", payload.PublishedAt)
    }

    post := database.PostsAttrs{
        AuthorID:    author.ID,
        Slug:        payload.Slug,
        Title:       payload.Title,
        Excerpt:     payload.Excerpt,
        Content:     payload.Content,
        PublishedAt: publishedAt,
        ImageURL:    payload.ImageURL,
        Author:      *author,
        Categories:  h.ParseCategories(payload),
        Tags:        h.ParseTags(payload),
    }

    fmt.Println("-----------------")
    fmt.Println(post)

    return nil
}

func (h *Handler) ParseCategories(payload *markdown.Post) []database.CategoriesAttrs {
    var categories []database.CategoriesAttrs

    slice := append(categories, database.CategoriesAttrs{
        Slug: payload.CategorySlug,
        Name: payload.Category,
    })

    return slice
}

func (h *Handler) ParseTags(payload *markdown.Post) []database.TagAttrs {
    var slice []database.TagAttrs

    for _, tag := range payload.Tags {
        slice = append(slice, database.TagAttrs{
            Slug: tag,
            Name: tag,
        })
    }

    return slice
}
