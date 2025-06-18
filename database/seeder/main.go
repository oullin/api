package main

import (
	"fmt"
	"github.com/oullin/boost"
	"github.com/oullin/database"
	"github.com/oullin/database/seeder/seeds"
	"github.com/oullin/env"
	"github.com/oullin/pkg/cli"
	"os"
	"os/exec"
	"sync"
	"time"
)

var environment *env.Environment

func init() {
	secrets, _ := boost.Spark("./.env")

	environment = secrets
}

func main() {
	clearScreen()
	dbConnection := boost.MakeDbConnection(environment)
	logs := boost.MakeLogs(environment)

	defer (*logs).Close()
	defer (*dbConnection).Close()

	// [1] --- Create the Seeder Runner.
	seeder := seeds.MakeSeeder(dbConnection, environment)

	// [2] --- Truncate the DB.
	if err := seeder.TruncateDB(); err != nil {
		panic(err)
	} else {
		cli.Successln("DB Truncated successfully ...")
		time.Sleep(2 * time.Second)
	}

	// [3] --- Seed users and posts sequentially because the below seeders depend on them.
	UserA, UserB := seeder.SeedUsers()
	posts := seeder.SeedPosts(UserA, UserB)

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

	// [4] Use channels to concurrently seed categories and tags since they are main dependencies.
	categories := <-categoriesChan
	tags := <-tagsChan

	// [5] Use a WaitGroup to run independent seeding tasks concurrently.
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
		seeder.SeedPostViews(posts, UserA, UserB)
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

	cli.Successln("DB seeded as expected.")
}

func clearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		message := fmt.Sprintf("Could not clear screen. Error: %s", err.Error())

		cli.Errorln(message)
	}
}
