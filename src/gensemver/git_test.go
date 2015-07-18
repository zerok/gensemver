package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRawLog(t *testing.T) {
	data := `commit 49f4b0894e578970e6397970a80856ad965603b4
tree fc88e7a33756cb1d863247d6b214a94fc9a84c3f
parent d39d5fee3fa97470cd551c6de3eb531fa6861f70
author Horst Gutmann <h.gutmann@netconomy.net> 1435652619 +0200
committer Horst Gutmann <h.gutmann@netconomy.net> 1435652619 +0200

    Title

commit 7a576ef53d7f0c43657ba378ce2127e1e85ba4bc
tree 14a7ac61bbbed5538b2142196f8ced094cf7ab66
parent a34d63bac735346a08063637dd409ff3e7dc4dcf
parent 6eaf34075904c4131aa5c9e2c84aaa957f48da1f
author Horst Gutmann <h.gutmann@netconomy.net> 1436949134 +0200
committer Horst Gutmann <h.gutmann@netconomy.net> 1436949134 +0200

    Title

    Body
`
	commits, err := ParseRawLog(data, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(commits))

	// Test first commit
	commit := commits[0]
	assert.Equal(t, Commit{
		ID:        "49f4b0894e578970e6397970a80856ad965603b4",
		TreeID:    "fc88e7a33756cb1d863247d6b214a94fc9a84c3f",
		ParentIDs: []string{"d39d5fee3fa97470cd551c6de3eb531fa6861f70"},
		Author:    "Horst Gutmann <h.gutmann@netconomy.net>",
		Committer: "Horst Gutmann <h.gutmann@netconomy.net>",
		MsgTitle:  "Title",
		MsgBody:   "",
		Tags:      nil}, commit)

	assert.Equal(t, "Title", commits[1].MsgTitle)
	assert.Equal(t, "Body", commits[1].MsgBody)

}
