package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"github.com/blang/semver"
)

type Tag struct {
	Name string
	Ref  string
}

func getTagsFromPackedRefs(repoPath string) []Tag {
	raw, err := ioutil.ReadFile(path.Join(repoPath, ".git", "packed-refs"))
	result := make([]Tag, 0, 10)
	if err != nil {
		if os.IsNotExist(err) {
			return result
		}
		return result
	}
	for _, line := range strings.Split(string(raw), "\n") {
		elems := strings.Split(line, " ")
		ref := elems[len(elems)-1]
		if strings.HasPrefix(ref, "refs/tags/") {
			tag := strings.TrimPrefix(ref, "refs/tags/")
			result = append(result, Tag{Name: tag, Ref: elems[0]})
		}
	}
	return result
}

func getTagsFromRefInfo(repoPath string) []Tag {
	raw, err := ioutil.ReadFile(path.Join(repoPath, ".git", "info", "refs"))
	result := make([]Tag, 0, 10)
	if err != nil {
		if os.IsNotExist(err) {
			return result
		}
		return result
	}
	for _, line := range strings.Split(string(raw), "\n") {
		elems := strings.Split(line, "\t")
		ref := elems[len(elems)-1]
		if strings.HasPrefix(ref, "refs/tags/") {
			tag := strings.TrimPrefix(ref, "refs/tags/")
			result = append(result, Tag{Name: tag, Ref: elems[0]})
		}
	}
	return result
}

func getTagsFromRefs(repoPath string) []Tag {
	tags := make([]Tag, 0, 10)
	entries, err := ioutil.ReadDir(path.Join(repoPath, ".git", "refs", "tags"))
	if err == nil {
		for _, entry := range entries {
			raw, err := ioutil.ReadFile(path.Join(repoPath, ".git", "refs", "tags", entry.Name()))
			if err != nil {
				log.Printf("Failed to parse reffile %s: %s", entry.Name(), err.Error())
				continue
			}
			tags = append(tags, Tag{Name: entry.Name(), Ref: strings.TrimSpace(string(raw))})
		}
	}
	return tags
}

func getTags(repoPath string) []Tag {
	tags := getTagsFromRefInfo(repoPath)
	for _, tag := range getTagsFromRefs(repoPath) {
		tags = append(tags, tag)
	}
	for _, tag := range getTagsFromPackedRefs(repoPath) {
		tags = append(tags, tag)
	}
	return tags
}

// getPreviousVersionRev checks the current repository for the latest version tag
// and returns its revision numbere
func getPreviousVersionRev(repoPath string, tags []Tag) (string, *semver.Version, error) {
	var previousVersion *semver.Version
	var previousRef string
	for _, tag := range tags {
		version, err := semver.Make(
			strings.TrimPrefix(strings.TrimPrefix(tag.Name, "R"), "v"))
		if err != nil {
			log.Printf("Failed to parse %s: %s", tag.Name, err.Error())
			continue
		}
		if previousVersion == nil || version.GT(*previousVersion) {
			previousVersion = &version
			previousRef = tag.Ref
		}
	}
	if previousVersion != nil {
		return previousRef, previousVersion, nil
	}
	return "", nil, nil
}

func getNewVersion(repoPath string, start, end string, oldVersion *semver.Version, tags []Tag) (*semver.Version, error) {
	commits, err := GetLog(start, end, tags)
	if err != nil {
		return nil, err
	}
	maxSev := -1
commitLoop:
	for _, commit := range commits {
		// Edge case: the last commit actually contains a version tag of the "old version"
		// and therefore should make this whole quest unnecessary. So if we find the tag
		// of the oldVersion in here, we have to reset the maxSev value.
		for _, version := range commit.GetVersions() {
			if version.Equals(*oldVersion) {
				maxSev = -1
				continue commitLoop
			}
		}

		sev, err := DetectChangeSeverity(&commit)
		if err != nil {
			continue
		}
		if sev > maxSev {
			maxSev = sev
		}
		if maxSev == MAJOR {
			break
		}
	}
	if maxSev == -1 {
		return nil, fmt.Errorf("No new version could be determined")
	}
	newVersion := IncrementVersion(*oldVersion, maxSev)
	return &newVersion, nil
}

func main() {
	customRawPrevVersion := flag.String("prev", "", "Force a previous version number")
	var customPrevVersion *semver.Version
	flag.Parse()

	if *customRawPrevVersion != "" {
		tmp := semver.MustParse(*customRawPrevVersion)
		customPrevVersion = &tmp
	}

	args := flag.Args()
	endRev := "HEAD"
	startRev := ""

	repoPath, err := GetRepoPath()
	tags := getTags(repoPath)

	var newVersion *semver.Version
	var oldVersion *semver.Version
	if err != nil {
		log.Fatalf("Failed to open repository: %s", err.Error())
	}
	if len(args) == 0 {
		startRev, oldVersion, err = getPreviousVersionRev(repoPath, tags)
		if err != nil {
			log.Fatalf("Failed to retrieve previous version: %s", err.Error())
		}
	} else if len(args) == 1 {
		_, oldVersion, err = getPreviousVersionRev(repoPath, tags)
		startRev = args[0]
	} else {
		startRev = args[0]
		_, oldVersion, err = getPreviousVersionRev(repoPath, tags)
		endRev = args[1]
	}

	if customPrevVersion != nil {
		oldVersion = customPrevVersion
	}

	if oldVersion == nil {
		log.Println("No previous version found")
		newVersion = &semver.Version{Major: 1, Minor: 0, Patch: 0}
	}

	if oldVersion != nil {
		log.Printf("Previous version: %s (%s)\n", oldVersion, startRev)
	}

	if newVersion == nil {
		newVersion, err = getNewVersion(repoPath, startRev, endRev, oldVersion, tags)
		if err != nil {
			log.Fatalf("Failed to generate new version: %s", err.Error())
		}
	}

	fmt.Println(newVersion)
}
