package strategies

import "app/model"

// PlayStrategy defines how next/previous songs are selected from a playlist
type PlayStrategy interface {
	SetPlaylist(playlist *model.Playlist)
	HasNext() bool
	Next() (*model.Song, error)
	HasPrevious() bool
	Previous() (*model.Song, error)
	AddToNext(song *model.Song) error // only meaningful for CustomQueue; others no-op
}
