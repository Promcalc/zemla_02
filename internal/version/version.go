package version

import "fmt"

var (
	// Эти переменные будут заполнены через ldflags при сборке
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
	goVersion = "unknown"
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
		Version:   version,
		Commit:    commit,
		BuildDate: buildDate,
		GoVersion: goVersion,
	}
}

// String возвращает строковое представление версии
func String() string {
	return fmt.Sprintf("lot-collector/%s (%s) built on %s with %s", 
		version, commit, buildDate, goVersion)
}