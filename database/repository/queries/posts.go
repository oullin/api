package queries

import (
	"gorm.io/gorm"
)

// ApplyPostsFilters The given query master table is "posts"
func ApplyPostsFilters(filters *PostFilters, query *gorm.DB) {
	if filters == nil {
		return
	}

	if filters.GetTitle() != "" {
		query.Where("LOWER(posts.title) ILIKE ?", "%"+filters.GetTitle()+"%")
	}

	if filters.GetText() != "" {
		query.
			Where("LOWER(posts.slug) ILIKE ? OR LOWER(posts.excerpt) ILIKE ? OR LOWER(posts.content) ILIKE ?",
				"%"+filters.GetText()+"%",
				"%"+filters.GetText()+"%",
				"%"+filters.GetText()+"%",
			)
	}

	if filters.GetAuthor() != "" {
		query.
			Joins("JOIN users ON posts.author_id = users.id").
			Where("users.deleted_at IS NULL").
			Where("("+
				"LOWER(users.bio) ILIKE ? OR LOWER(users.first_name) ILIKE ? OR LOWER(users.last_name) ILIKE ? OR LOWER(users.display_name) ILIKE ?"+
				")",
				"%"+filters.GetAuthor()+"%",
				"%"+filters.GetAuthor()+"%",
				"%"+filters.GetAuthor()+"%",
				"%"+filters.GetAuthor()+"%",
			)
	}

	if filters.GetCategory() != "" {
		query.
			Joins("JOIN post_categories ON post_categories.post_id = posts.id").
			Joins("JOIN categories ON categories.id = post_categories.category_id").
			Where("categories.deleted_at IS NULL").
			Where("("+
				"LOWER(categories.slug) ILIKE ? OR LOWER(categories.name) ILIKE ? OR LOWER(categories.description) ILIKE ?"+
				")",
				"%"+filters.GetCategory()+"%",
				"%"+filters.GetCategory()+"%",
				"%"+filters.GetCategory()+"%",
			)
	}

	if filters.GetTag() != "" {
		query.
			Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Joins("JOIN tags ON tags.id = post_tags.tag_id").
			Where("tags.deleted_at IS NULL").
			Where("("+
				"LOWER(tags.slug) ILIKE ? OR LOWER(tags.name) ILIKE ? OR LOWER(tags.description) ILIKE ?"+
				")",
				"%"+filters.GetTag()+"%",
				"%"+filters.GetTag()+"%",
				"%"+filters.GetTag()+"%",
			)
	}
}
