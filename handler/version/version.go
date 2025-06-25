package version

// Section represents a specific section of a user's profile or portfolio.
type Section string

const (
	ProfileSection    Section = "profile"
	ExperienceSection Section = "experience"
	ProjectsSection   Section = "projects"
	SocialSection     Section = "social"
	TalksSection      Section = "talks"
)

func getSectionVersion(section Section) string {
	versionMap := map[Section]string{
		ProfileSection:    "v1.2.0",
		ExperienceSection: "v1.1.0",
		ProjectsSection:   "v2.0.1",
		SocialSection:     "v1.0.0",
		TalksSection:      "v1.3.2",
	}

	version, ok := versionMap[section]
	if !ok {
		return "unknown"
	}
	return version
}
