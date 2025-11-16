package version

import "fmt"

// Эти переменные будут заполнены через ldflags при сборке
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
	GoVersion = "unknown"
)

// VersionInfo содержит информацию о версии приложения
type VersionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
}

// Get возвращает информацию о версии
func Get() VersionInfo {
	return VersionInfo{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
	}
}

// String возвращает строковое представление версии
func String() string {
	return fmt.Sprintf("lot-collector/%s (%s) built on %s with %s",
		Version, Commit, BuildDate, GoVersion)
}
