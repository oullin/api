package router

import (
	"fmt"
)

const FixtureTalks = "talks"
const FixtureSocial = "social"
const FixtureProfile = "profile"
const FixtureProjects = "projects"
const FixtureEducation = "education"
const FixtureExperience = "experience"
const FixtureRecommendations = "recommendations"

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
	return f.resolveFor(FixtureTalks)
}

func (f *Fixture) GetTalksFile() string {
	return f.resolveFor(FixtureTalks).fullPath

}

func (f *Fixture) GetSocial() *Fixture {
	return f.resolveFor(FixtureSocial)
}

func (f *Fixture) GetSocialFile() string {
	return f.resolveFor(FixtureSocial).fullPath
}

func (f *Fixture) GetProfile() *Fixture {
	return f.resolveFor(FixtureProfile)
}

func (f *Fixture) GetProfileFile() string {
	return f.resolveFor(FixtureProfile).fullPath
}

func (f *Fixture) GetProjects() *Fixture {
	return f.resolveFor(FixtureProjects)
}

func (f *Fixture) GetProjectsFile() string {
	return f.resolveFor(FixtureProjects).fullPath
}

func (f *Fixture) GetEducation() *Fixture {
	return f.resolveFor(FixtureEducation)
}

func (f *Fixture) GetEducationFile() string {
	return f.resolveFor(FixtureEducation).fullPath
}

func (f *Fixture) GetExperience() *Fixture {
	return f.resolveFor(FixtureExperience)
}

func (f *Fixture) GetExperienceFile() string {
	return f.resolveFor(FixtureExperience).fullPath
}

func (f *Fixture) GetRecommendations() *Fixture {
	return f.resolveFor(FixtureRecommendations)
}

func (f *Fixture) GetRecommendationsFile() string {
	return f.resolveFor(FixtureRecommendations).fullPath
}

func (f *Fixture) resolveFor(slug string) *Fixture {
	clone := f
	clone.fullPath = clone.getFileFor(slug)
	clone.file = slug

	return clone
}

func (f *Fixture) getFileFor(slug string) string {
	return fmt.Sprintf("%s%s.%s", f.basePath, slug, f.mime)
}
