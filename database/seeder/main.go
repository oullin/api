package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/oullin/database"
	"github.com/oullin/database/seeder/seeds"
	"github.com/oullin/metal/kernel"
	"github.com/oullin/pkg/cli"
	"github.com/oullin/pkg/portal"
)

func main() {
	if err := run(); err != nil {
		sentry.CurrentHub().CaptureException(err)
		cli.Errorln(err.Error())
		os.Exit(1)
	}
}

func run() error {
	cli.ClearScreen()

	validate := portal.GetDefaultValidator()
	environment := kernel.Ignite("./.env", validate)
	if environment == nil {
		return errors.New("environment is nil")
	}

	hub := kernel.NewSentry(environment)

	defer sentry.Flush(2 * time.Second)
	defer kernel.RecoverWithSentry(hub)

	dbConnection := kernel.NewDbConnection(environment)
	if dbConnection == nil {
		return errors.New("database connection is nil")
	}
	defer dbConnection.Close()

	logs := kernel.NewLogs(environment)
	if logs == nil {
		return errors.New("logs driver is nil")
	}
	defer logs.Close()

	seeder := seeds.NewSeeder(dbConnection, environment)
	if seeder == nil {
		return errors.New("seeder is nil")
	}

	if err := seeder.TruncateDB(); err != nil {
		return fmt.Errorf("truncate database: %w", err)
	}

	cli.Successln("db Truncated successfully ...")
	time.Sleep(2 * time.Second)

	userA, userB := seeder.SeedUsers()
	posts := seeder.SeedPosts(userA, userB)

	categoriesChan := make(chan []database.Category)
	tagsChan := make(chan []database.Tag)

	go func() {
		defer close(categoriesChan)

		cli.Warningln("Seeding categories ...")
		categoriesChan <- seeder.SeedCategories()
	}()

	go func() {
		defer close(tagsChan)

		cli.Magentaln("Seeding tags ...")
		tagsChan <- seeder.SeedTags()
	}()

	categories := <-categoriesChan
	tags := <-tagsChan

	var wg sync.WaitGroup
	wg.Add(6)

	go func() {
		defer wg.Done()

		cli.Blueln("Seeding comments ...")
		seeder.SeedComments(posts...)
	}()

	go func() {
		defer wg.Done()

		cli.Cyanln("Seeding likes ...")
		seeder.SeedLikes(posts...)
	}()

	go func() {
		defer wg.Done()

		cli.Grayln("Seeding posts-categories ...")
		seeder.SeedPostsCategories(categories, posts)
	}()

	go func() {
		defer wg.Done()

		cli.Grayln("Seeding posts-tags ...")
		seeder.SeedPostTags(tags, posts)
	}()

	go func() {
		defer wg.Done()

		cli.Warningln("Seeding views ...")
		seeder.SeedPostViews(posts, userA, userB)
	}()

	go func() {
		defer wg.Done()

		if err := seeder.SeedNewsLetters(); err != nil {
			cli.Error(err.Error())
		} else {
			cli.Successln("Seeding Newsletters ...")
		}
	}()

	wg.Wait()

	cli.Magentaln("db seeded as expected ....")

	return nil
}
