package seeds

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/metal/env"
	"math/rand"
	"time"
)

type Seeder struct {
	dbConn      *database.Connection
	environment *env.Environment
}

func MakeSeeder(dbConnection *database.Connection, environment *env.Environment) *Seeder {
	return &Seeder{
		dbConn:      dbConnection,
		environment: environment,
	}
}

func (s *Seeder) TruncateDB() error {
	if s.environment.App.IsProduction() {
		return fmt.Errorf("cannot truncate db at the seeder level")
	}

	truncate := database.MakeTruncate(s.dbConn, s.environment)

	if err := truncate.Execute(); err != nil {
		panic(err)
	}

	return nil
}

func (s *Seeder) SeedUsers() (database.User, database.User) {
	users := MakeUsersSeed(s.dbConn)

	UserA, err := users.Create(database.UsersAttrs{
		Username: "gocanto",
		Name:     "Gus",
		IsAdmin:  true,
	})

	if err != nil {
		panic(err)
	}

	UserB, err := users.Create(database.UsersAttrs{
		Username: "li",
		Name:     "liane",
		IsAdmin:  false,
	})

	if err != nil {
		panic(err)
	}

	return UserA, UserB
}

func (s *Seeder) SeedPosts(UserA, UserB database.User) []database.Post {
	posts := MakePostsSeed(s.dbConn)
	timex := time.Now()

	PostsA, err := posts.CreatePosts(database.PostsAttrs{
		AuthorID:    UserA.ID,
		Slug:        fmt.Sprintf("post-slug-%s", uuid.NewString()),
		Title:       fmt.Sprintf("Post %s title", uuid.NewString()),
		Excerpt:     fmt.Sprintf("[%s] Sed at risus vel nulla consequat fermentum. Donec et orci mauris", uuid.NewString()),
		Content:     fmt.Sprintf("[%s] Sed at risus vel nulla consequat fermentum. Donec et orci mauris. Nullam tempor velit id mi luctus, a scelerisque libero accumsan. In hac habitasse platea dictumst. Cras ac nunc nec massa tristique fringilla.", uuid.NewString()),
		PublishedAt: &timex,
	}, 1)

	if err != nil {
		panic(err)
	}

	PostsB, err := posts.CreatePosts(database.PostsAttrs{
		AuthorID:    UserB.ID,
		Slug:        fmt.Sprintf("post-slug-%s", uuid.NewString()),
		Title:       fmt.Sprintf("Post %s title", uuid.NewString()),
		Excerpt:     fmt.Sprintf("[%s] Sed at risus vel nulla consequat fermentum. Donec et orci mauris", uuid.NewString()),
		Content:     fmt.Sprintf("[%s] Sed at risus vel nulla consequat fermentum. Donec et orci mauris. Nullam tempor velit id mi luctus, a scelerisque libero accumsan. In hac habitasse platea dictumst. Cras ac nunc nec massa tristique fringilla.", uuid.NewString()),
		PublishedAt: &timex,
	}, 1)

	if err != nil {
		panic(err)
	}

	return append(PostsA, PostsB...)
}

func (s *Seeder) SeedCategories() []database.Category {
	categories := MakeCategoriesSeed(s.dbConn)

	result, err := categories.Create(database.CategoriesAttrs{
		Slug:        fmt.Sprintf("category-slug-%s", uuid.NewString()),
		Name:        fmt.Sprintf("category-slug-%s", uuid.NewString()),
		Description: fmt.Sprintf("[%s] Sed at risus vel nulla consequat fermentum. Donec et orci mauris", uuid.NewString()),
	})

	if err != nil {
		panic(err)
	}

	return result
}

func (s *Seeder) SeedTags() []database.Tag {
	seed := MakeTagsSeed(s.dbConn)

	tags, err := seed.Create()

	if err != nil {
		panic(err)
	}

	return tags
}

func (s *Seeder) SeedComments(posts ...database.Post) {
	seed := MakeCommentsSeed(s.dbConn)

	timex := time.Now()
	var values []database.CommentsAttrs

	for index, post := range posts {
		values = append(values, database.CommentsAttrs{
			PostID:     post.ID,
			AuthorID:   post.AuthorID,
			ParentID:   nil,
			Content:    fmt.Sprintf("[%d] Nullam quis arcu in magna pulvinar tincidunt. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nam hendrerit nulla ut cursus laoreet. Nullam elementum lorem vel facilisis laoreet. Cras ac turpis vel erat vehicula venenatis.", index),
			ApprovedAt: &timex,
		})
	}

	if _, err := seed.Create(values...); err != nil {
		panic(err)
	}
}

func (s *Seeder) SeedLikes(posts ...database.Post) {
	seed := MakeLikesSeed(s.dbConn)
	var values []database.LikesAttrs

	for _, post := range posts {
		values = append(values, database.LikesAttrs{
			PostID: post.ID,
			UserID: post.AuthorID,
		})
	}

	_, err := seed.Create(values...)

	if err != nil {
		panic(err)
	}
}

func (s *Seeder) SeedPostsCategories(categories []database.Category, posts []database.Post) {
	if len(categories) == 0 || len(posts) == 0 {
		return
	}

	seed := MakePostCategoriesSeed(s.dbConn)

	var post database.Post
	var category database.Category

	source := rand.NewSource(time.Now().UnixNano())
	salt := rand.New(source)

	cIndex := salt.Intn(len(categories))
	category = categories[cIndex]

	pIndex := salt.Intn(len(posts))
	post = posts[pIndex]

	err := seed.Create(category, post)

	if err != nil {
		panic(err)
	}
}

func (s *Seeder) SeedPostTags(tags []database.Tag, posts []database.Post) {
	if len(tags) == 0 || len(posts) == 0 {
		return
	}

	seed := MakePostTagsSeed(s.dbConn)

	var post database.Post
	var label database.Tag

	source := rand.NewSource(time.Now().UnixNano())
	salt := rand.New(source)

	tIndex := salt.Intn(len(tags))
	label = tags[tIndex]

	pIndex := salt.Intn(len(posts))
	post = posts[pIndex]

	err := seed.Create(label, post)

	if err != nil {
		panic(err)
	}
}

func (s *Seeder) SeedPostViews(posts []database.Post, users ...database.User) {
	if len(posts) == 0 || len(users) == 0 {
		return
	}

	seed := MakePostViewsSeed(s.dbConn)

	var values []database.PostViewsAttr

	for pIndex, post := range posts {
		for uIndex, user := range users {
			values = append(values, database.PostViewsAttr{
				Post:      post,
				User:      user,
				IPAddress: fmt.Sprintf("192.168.0.%d", pIndex+1),
				UserAgent: fmt.Sprintf("[post:%d][user:%d] - Mozilla/5.0 (Macintosh; ...) ...", pIndex+1, uIndex+1),
			})
		}
	}

	err := seed.Create(values)

	if err != nil {
		panic(err)
	}
}

func (s *Seeder) SeedNewsLetters() error {
	var newsletters []database.NewsletterAttrs

	a := database.NewsletterAttrs{
		FirstName:      "John",
		LastName:       "Smith",
		Email:          "john.smith@gmail.com",
		SubscribedAt:   nil,
		UnsubscribedAt: nil,
	}

	currentTime := time.Now()
	last3Month := currentTime.AddDate(0, -3, 0)
	b := database.NewsletterAttrs{
		FirstName:      "Don",
		LastName:       "Smith",
		Email:          "Don.smith@gmail.com",
		SubscribedAt:   &last3Month,
		UnsubscribedAt: &currentTime,
	}

	seed := MakeNewslettersSeed(s.dbConn)
	newsletters = append(newsletters, a, b)

	if err := seed.Create(newsletters); err != nil {
		panic(err)
	}

	return nil
}
