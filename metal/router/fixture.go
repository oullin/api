package router

import (
	"fmt"
)

const fixtureTalks = "talks"
const fixtureSocial = "social"
const fixtureProfile = "profile"
const fixtureProjects = "projects"
const fixtureEducation = "education"
const fixtureExperience = "experience"
const fixtureRecommendations = "recommendations"

type Fixture struct {
	file     string
	basePath string
	fullPath string
	mime     string
}

func NewFixture() Fixture {
	return Fixture{
		basePath: "./storage/fixture/",
		mime:     "json",
	}
}

func (f *Fixture) GetTalks() *Fixture {
	return f.resolveFor(fixtureTalks)
}

func (f *Fixture) GetSocial() *Fixture {
	return f.resolveFor(fixtureSocial)
}

func (f *Fixture) GetProfile() *Fixture {
	return f.resolveFor(fixtureProfile)
}

func (f *Fixture) GetProjects() *Fixture {
	return f.resolveFor(fixtureProjects)
}

func (f *Fixture) GetEducation() *Fixture {
	return f.resolveFor(fixtureEducation)
}

func (f *Fixture) GetExperience() *Fixture {
	return f.resolveFor(fixtureExperience)
}

func (f *Fixture) GetRecommendations() *Fixture {
	return f.resolveFor(fixtureRecommendations)
}

func (f *Fixture) resolveFor(slug string) *Fixture {
	f.fullPath = f.getFileFor(slug)
	f.file = slug

	return f
}

func (f *Fixture) getFileFor(slug string) string {
	return fmt.Sprintf("%s%s.%s", f.basePath, slug, f.mime)
}
