package main

import (
	"fmt"
	"strings"

	"github.com/blang/semver"
)

const (
	PATCH = iota
	MINOR = iota
	MAJOR = iota
)

// DetectChangeSeverity parses the title and the body of a commit in order
// to determine if the change it contains affects the current version on the
// major, minor, or patch level. If no markers indicating that can be found
// an error is returned
func DetectChangeSeverity(commit *Commit) (int, error) {
	patchPrefixes := []string{"docs", "fix", "chore", "style", "test"}

	if strings.Contains(commit.MsgBody, "BREAKING CHANGES:") {
		return MAJOR, nil
	}

	if strings.HasPrefix(commit.MsgTitle, "feat(") || strings.HasPrefix(commit.MsgTitle, "feat:") {
		return MINOR, nil
	}

	for _, patchPrefix := range patchPrefixes {
		if strings.HasPrefix(commit.MsgTitle, patchPrefix+"(") || strings.HasPrefix(commit.MsgTitle, patchPrefix+":") {
			return PATCH, nil
		}
	}

	return 0, fmt.Errorf("Commit doesn't follow the commit msg rules")
}

// IncrementVersion generates a new version based on the previous version
// number and the given severity level. Note that if the current release
// is a pre-release, it will only increment that counter.
func IncrementVersion(prev semver.Version, severity int) semver.Version {
	result := prev
	if severity == PATCH {
		result.Patch += 1
	} else if severity == MINOR {
		result.Minor += 1
		result.Patch = 0
	} else if severity == MAJOR {
		result.Major += 1
		result.Minor = 0
		result.Patch = 0
	}
	return result
}
