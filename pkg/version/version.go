package version

var (
	buildVersion = ""
	buildTime    = ""
)

// GetBuildVersion returns the foobar-operator build version
func GetBuildVersion() string {
	return buildVersion
}

// GetBuildTime returns the foobar-operator build time
func GetBuildTime() string {
	return buildTime
}
