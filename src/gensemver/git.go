package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
)

type Commit struct {
	ID        string
	TreeID    string
	ParentIDs []string
	Author    string
	Committer string
	MsgTitle  string
	MsgBody   string
	Tags      []string
}

func (c Commit) GetVersions() []semver.Version {
	result := make([]semver.Version, 0, 0)
	if c.Tags == nil || len(c.Tags) == 0 {
		return result
	}
	for _, tag := range c.Tags {
		ver, err := ExtractVersion(tag)
		if err == nil {
			result = append(result, ver)
		}
	}
	return result
}

func ExtractVersion(tagname string) (semver.Version, error) {
	strippedTag := strings.TrimPrefix(strings.TrimPrefix(tagname, "R"), "v")
	return semver.Make(strippedTag)
}

func ParseRawLog(log string, tags []Tag) ([]Commit, error) {
	tagMap := make(map[string][]string)
	if tags != nil {
		for _, tag := range tags {
			prev, ok := tagMap[tag.Ref]
			if !ok {
				prev = make([]string, 0, 1)
			}
			prev = append(prev, tag.Name)
			tagMap[tag.Ref] = prev
		}
	}
	commits := make([]Commit, 0, 10)
	var commit *Commit
	msgBody := make([]string, 0)
	for _, line := range strings.Split(log, "\n") {
		if strings.HasPrefix(line, "commit ") {
			if commit != nil {
				commit.MsgBody = strings.Join(msgBody, "\n")
				commits = append(commits, *commit)
				msgBody = make([]string, 0)
			}
			commitID := strings.TrimPrefix(line, "commit ")
			commit = &Commit{ID: commitID, ParentIDs: make([]string, 0, 1), Tags: nil}
			foundTags, ok := tagMap[commitID]
			if ok {
				commit.Tags = foundTags
			}
		} else if strings.HasPrefix(line, "author ") {
			commit.Author = stripAuthorTimestamp(strings.TrimPrefix(line, "author "))
		} else if strings.HasPrefix(line, "committer ") {
			commit.Committer = stripAuthorTimestamp(strings.TrimPrefix(line, "committer "))
		} else if strings.HasPrefix(line, "tree ") {
			commit.TreeID = strings.TrimPrefix(line, "tree ")
		} else if strings.HasPrefix(line, "parent ") {
			commit.ParentIDs = append(commit.ParentIDs, strings.TrimPrefix(line, "parent "))
		} else {
			if commit == nil {
				continue
			}
			contentLine := trimMsgLine(line)
			if commit.MsgTitle == "" {
				commit.MsgTitle = contentLine
			} else if contentLine == "" && commit.MsgBody == "" {
				// We are in the separator right now and so we can move on.
				continue
			} else {
				msgBody = append(msgBody, contentLine)
			}
		}
	}
	if commit != nil {
		commit.MsgBody = strings.Join(msgBody, "\n")
		commits = append(commits, *commit)
	}
	return commits, nil
}

// GetRepoPath returns the nearest directory from CWD that contains a .git folder.
func GetRepoPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", nil
	}
	for {
		_, err := os.Stat(filepath.Join(cwd, ".git"))
		if err != nil {
			if !os.IsNotExist(err) {
				return "", err
			}
			newCwd := filepath.Dir(cwd)
			if newCwd == cwd {
				return "", fmt.Errorf("You are not in a git repository!")
			}
			cwd = newCwd
		} else {
			return cwd, nil
		}
	}
}

func GetLog(start, end string, tags []Tag) ([]Commit, error) {
	executable, err := exec.LookPath("git")
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(executable,
		"log",
		"--reverse",
		"--pretty=raw",
		fmt.Sprintf("%s..%s", start, end),
	)
	raw, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	return ParseRawLog(string(raw), tags)
}

func stripAuthorTimestamp(line string) string {
	elems := strings.Split(line, " ")
	return strings.Join(elems[0:len(elems)-2], " ")
}

func trimMsgLine(line string) string {
	if strings.HasPrefix(line, "\t") {
		return strings.TrimPrefix(line, "\t")
	} else if strings.HasPrefix(line, "    ") {
		return strings.TrimPrefix(line, "    ")
	}
	return line
}
