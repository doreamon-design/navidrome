package core

import (
	"github.com/doreamon-design/navidrome/core/agents"
	"github.com/doreamon-design/navidrome/core/ffmpeg"
	"github.com/doreamon-design/navidrome/core/scrobbler"
	"github.com/google/wire"
)

var Set = wire.NewSet(
	NewMediaStreamer,
	GetTranscodingCache,
	NewArchiver,
	NewExternalMetadata,
	NewPlayers,
	NewShare,
	NewPlaylists,
	agents.New,
	ffmpeg.New,
	scrobbler.GetPlayTracker,
)
