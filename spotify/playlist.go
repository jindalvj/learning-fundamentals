package models

import "errors"

type Playlist struct {
	Name  string
	songs []*Song
}

func NewPlaylist(name string) *Playlist {
	return &Playlist{Name: name}
}

func (p *Playlist) AddSong(song *Song) error {
	if song == nil {
		return errors.New("cannot add nil song to playlist")
	}
	p.songs = append(p.songs, song)
	return nil
}

func (p *Playlist) Songs() []*Song {
	return p.songs
}

func (p *Playlist) Size() int {
	return len(p.songs)
}
