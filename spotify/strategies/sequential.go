package strategies

import (
	"app/model"
	"errors"
)

type SequentialPlayStrategy struct {
	playlist     *model.Playlist
	currentIndex int
}

func NewSequentialPlayStrategy() *SequentialPlayStrategy {
	return &SequentialPlayStrategy{currentIndex: -1}
}

func (s *SequentialPlayStrategy) SetPlaylist(playlist *model.Playlist) {
	s.playlist = playlist
	s.currentIndex = -1
}

func (s *SequentialPlayStrategy) HasNext() bool {
	if s.playlist == nil {
		return false
	}
	return s.currentIndex+1 < s.playlist.Size()
}

func (s *SequentialPlayStrategy) Next() (*model.Song, error) {
	if s.playlist == nil || s.playlist.Size() == 0 {
		return nil, errors.New("no playlist loaded or playlist is empty")
	}
	if !s.HasNext() {
		return nil, errors.New("no more songs in playlist")
	}
	s.currentIndex++
	return s.playlist.Songs()[s.currentIndex], nil
}

func (s *SequentialPlayStrategy) HasPrevious() bool {
	return s.currentIndex-1 >= 0
}

func (s *SequentialPlayStrategy) Previous() (*model.Song, error) {
	if s.playlist == nil || s.playlist.Size() == 0 {
		return nil, errors.New("no playlist loaded or playlist is empty")
	}
	if !s.HasPrevious() {
		return nil, errors.New("already at the first song")
	}
	s.currentIndex--
	return s.playlist.Songs()[s.currentIndex], nil
}

func (s *SequentialPlayStrategy) AddToNext(_ *model.Song) error {
	return nil // no-op for sequential
}
