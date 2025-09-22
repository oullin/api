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
	basePath string
	fileType string
}

func NewFixture() Fixture {
	return Fixture{
		basePath: "./storage/fixture/",
		fileType: "json",
	}
}

func (f Fixture) GetTalks() string {
	return f.GetFileFor(fixtureTalks)
}

func (f Fixture) GetSocial() string {
	return f.GetFileFor(fixtureSocial)
}

func (f Fixture) GetProfile() string {
	return f.GetFileFor(fixtureProfile)
}

func (f Fixture) GetProjects() string {
	return f.GetFileFor(fixtureProjects)
}

func (f Fixture) GetEducation() string {
	return f.GetFileFor(fixtureEducation)
}

func (f Fixture) GetExperience() string {
	return f.GetFileFor(fixtureExperience)
}

func (f Fixture) GetRecommendations() string {
	return f.GetFileFor(fixtureRecommendations)
}

func (f Fixture) GetFileFor(slug string) string {
	return fmt.Sprintf("%s%s.%s", f.basePath, slug, f.fileType)
}
