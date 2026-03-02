package strategies

import (
	"errors"
	"math/rand"
	"musicplayer/models"
)

type RandomPlayStrategy struct {
	playlist      *models.Playlist
	remaining     []*models.Song
	history       []*models.Song
}

func NewRandomPlayStrategy() *RandomPlayStrategy {
	return &RandomPlayStrategy{}
}

func (r *RandomPlayStrategy) SetPlaylist(playlist *models.Playlist) {
	r.playlist = playlist
	r.history = nil
	// Copy songs into remaining pool
	r.remaining = make([]*models.Song, len(playlist.Songs()))
	copy(r.remaining, playlist.Songs())
}

func (r *RandomPlayStrategy) HasNext() bool {
	return len(r.remaining) > 0
}

func (r *RandomPlayStrategy) Next() (*models.Song, error) {
	if r.playlist == nil || r.playlist.Size() == 0 {
		return nil, errors.New("no playlist loaded or playlist is empty")
	}
	if len(r.remaining) == 0 {
		return nil, errors.New("no songs left to play")
	}

	idx := rand.Intn(len(r.remaining))
	selected := r.remaining[idx]

	// Swap and pop for O(1) removal
	last := len(r.remaining) - 1
	r.remaining[idx] = r.remaining[last]
	r.remaining = r.remaining[:last]

	r.history = append(r.history, selected)
	return selected, nil
}

func (r *RandomPlayStrategy) HasPrevious() bool {
	return len(r.history) > 0
}

func (r *RandomPlayStrategy) Previous() (*models.Song, error) {
	if len(r.history) == 0 {
		return nil, errors.New("no previous song available")
	}
	last := len(r.history) - 1
	song := r.history[last]
	r.history = r.history[:last]
	return song, nil
}

func (r *RandomPlayStrategy) AddToNext(_ *models.Song) error {
	return nil // no-op for random
}
