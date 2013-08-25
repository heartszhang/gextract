package feeds

import (
	"testing"
)

func TestInsertChannel(t *testing.T) {
	c := Channel{Title: "channel-one"}
	InsertChannel(c)
}

func TestTopNEntries(t *testing.T) {
	entries := TopNEntries(0, 2)
	t.Log("entries", len(entries))
}

func TestRefreshChannels(t *testing.T) {
	chs := RefreshChannels()
	t.Log("channels", len(chs))
}
