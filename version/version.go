package version

var (
	version string
	build   string
)

// GetVersion returns a version as string.
func GetVersion() string {
	return version
}

// GetBuild returns a build date as string.
func GetBuild() string {
	return build
}
