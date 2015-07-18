package main

import (
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
)

type SeverityTestCase struct {
	expectedErr      error
	expectedSeverity int
	commit           Commit
}

type IncrementVersionTestCase struct {
	expected semver.Version
	prev     semver.Version
	severity int
}

func TestDetectChangeSeverity(t *testing.T) {
	tests := []SeverityTestCase{
		SeverityTestCase{expectedErr: nil, expectedSeverity: MINOR, commit: Commit{
			MsgTitle: "feat(...): adsf",
		}},
		SeverityTestCase{expectedErr: nil, expectedSeverity: PATCH, commit: Commit{
			MsgTitle: "docs: adsf",
		}},
		SeverityTestCase{expectedErr: nil, expectedSeverity: MAJOR, commit: Commit{
			MsgTitle: "feat(...): adsf",
			MsgBody:  "BREAKING CHANGES:",
		}},
	}
	for _, testcase := range tests {
		severity, err := DetectChangeSeverity(&testcase.commit)
		assert.Equal(t, testcase.expectedErr, err, "%v", testcase.commit)
		assert.Equal(t, testcase.expectedSeverity, severity, "%v", testcase.commit)
	}
}

func TestIncrementVersion(t *testing.T) {
	tests := []IncrementVersionTestCase{
		IncrementVersionTestCase{prev: semver.Version{
			Major: 1,
			Minor: 1,
			Patch: 0,
		}, severity: MAJOR, expected: semver.Version{
			Major: 2,
			Minor: 0,
			Patch: 0,
		}},
		IncrementVersionTestCase{prev: semver.Version{
			Major: 1,
			Minor: 1,
			Patch: 0,
		}, severity: MINOR, expected: semver.Version{
			Major: 1,
			Minor: 2,
			Patch: 0,
		}},
		IncrementVersionTestCase{prev: semver.Version{
			Major: 1,
			Minor: 1,
			Patch: 0,
		}, severity: PATCH, expected: semver.Version{
			Major: 1,
			Minor: 1,
			Patch: 1,
		}},
	}
	for _, test := range tests {
		result := IncrementVersion(test.prev, test.severity)
		assert.Equal(t, test.expected, result)
	}
}
