package version

var (
	version string
	build   string
)

// Get returns a version string
func Get() string {
	return version
}
