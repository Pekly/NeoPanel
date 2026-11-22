// Package serverjar inspects server artifacts without executing their contents.
package serverjar

import "time"

// Info distinguishes configured settings from evidence about the current bytes.
// MinimumJava is bytecode evidence, not a claim of a supported runtime.
type Info struct {
	Status           string    `json:"status"`
	Source           string    `json:"source,omitempty"`
	Software         string    `json:"software,omitempty"`
	MinecraftVersion string    `json:"minecraft_version,omitempty"`
	ProjectVersion   string    `json:"project_version,omitempty"`
	BuildID          string    `json:"build_id,omitempty"`
	BuildName        string    `json:"build_name,omitempty"`
	RequiredJava     int       `json:"required_java,omitempty"`
	JavaSource       string    `json:"java_source,omitempty"`
	MinimumJava      int       `json:"minimum_java,omitempty"`
	SHA256           string    `json:"sha256,omitempty"`
	Filename         string    `json:"filename,omitempty"`
	Size             int64     `json:"size,omitempty"`
	Modified         time.Time `json:"modified,omitempty"`
	Message          string    `json:"message,omitempty"`
}
