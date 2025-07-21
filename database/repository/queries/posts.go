package queries

import (
	"gorm.io/gorm"
)

func ApplyPostsFilters(filters *PostFilters, query *gorm.DB) {
	if filters == nil {
		return
	}

	if filters.GetTitle() != "" {
		query.Where("LOWER(posts.title) ILIKE ?", "%"+filters.GetTitle()+"%")
	}

	if filters.GetText() != "" {
		query.
			Or("LOWER(posts.slug) ILIKE ?", "%"+filters.GetText()+"%").
			Or("LOWER(posts.excerpt) ILIKE ?", "%"+filters.GetText()+"%").
			Or("LOWER(posts.content) ILIKE ?", "%"+filters.GetText()+"%")
	}

	if filters.GetAuthor() != "" {
		query.
			Joins("Author").
			Where("LOWER(Author.username) = ?", filters.GetAuthor()).
			Or("LOWER(Author.first_name) = ?", filters.GetAuthor()).
			Or("LOWER(Author.last_name) = ?", filters.GetAuthor()).
			Or("LOWER(Author.display_name) = ?", filters.GetAuthor())
	}

	if filters.GetCategory() != "" {
		query.
			Joins("JOIN post_categories ON post_categories.post_id = posts.id").
			Joins("JOIN categories ON categories.id = post_categories.category_id").
			Where("LOWER(categories.slug) = ?", filters.GetCategory()).
			Or("LOWER(categories.name) = ?", filters.GetCategory()).
			Or("LOWER(categories.description) = ?", "%"+filters.GetCategory()+"%")
	}

	if filters.GetTag() != "" {
		query.
			Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Joins("JOIN tags ON tags.id = post_tags.tag_id").
			Where("LOWER(tags.slug) = ?", filters.GetTag()).
			Or("LOWER(tags.name) = ?", filters.GetTag()).
			Or("LOWER(tags.description) = ?", "%"+filters.GetTag()+"%")
	}
}
