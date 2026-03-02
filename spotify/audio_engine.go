package core

import (
	"errors"
	"fmt"
	"musicplayer/device"
	"musicplayer/models"
)

type AudioEngine struct {
	currentSong *models.Song
	paused      bool
}

func NewAudioEngine() *AudioEngine {
	return &AudioEngine{}
}

func (a *AudioEngine) CurrentSongTitle() string {
	if a.currentSong != nil {
		return a.currentSong.Title
	}
	return ""
}

func (a *AudioEngine) IsPaused() bool {
	return a.paused
}

func (a *AudioEngine) Play(output device.IAudioOutputDevice, song *models.Song) error {
	if song == nil {
		return errors.New("cannot play a nil song")
	}

	// Resume if same song was paused
	if a.paused && a.currentSong == song {
		a.paused = false
		fmt.Printf("Resuming song: %s\n", song.Title)
		output.PlayAudio(song)
		return nil
	}

	a.currentSong = song
	a.paused = false
	fmt.Printf("Playing song: %s\n", song.Title)
	output.PlayAudio(song)
	return nil
}

func (a *AudioEngine) Pause() error {
	if a.currentSong == nil {
		return errors.New("no song is currently playing")
	}
	if a.paused {
		return errors.New("song is already paused")
	}
	a.paused = true
	fmt.Printf("Pausing song: %s\n", a.currentSong.Title)
	return nil
}
