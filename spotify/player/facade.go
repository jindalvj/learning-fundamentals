package player

import (
	"app/core"
	"app/enum"
	"app/managers"
	"app/model"
	"app/strategies"
	"errors"
	"fmt"
	"sync"
)

type MusicPlayerFacade struct {
	audioEngine    *core.AudioEngine
	loadedPlaylist *model.Playlist
	playStrategy   strategies.PlayStrategy
}

var (
	facadeOnce     sync.Once
	facadeInstance *MusicPlayerFacade
)

func GetMusicPlayerFacade() *MusicPlayerFacade {
	facadeOnce.Do(func() {
		facadeInstance = &MusicPlayerFacade{
			audioEngine: core.NewAudioEngine(),
		}
	})
	return facadeInstance
}

func (f *MusicPlayerFacade) ConnectDevice(deviceType enum.DeviceType) {
	managers.GetDeviceManager().Connect(deviceType)
}

func (f *MusicPlayerFacade) SetPlayStrategy(strategyType enum.PlayStrategyType) error {
	s, err := managers.GetStrategyManager().GetStrategy(strategyType)
	if err != nil {
		return err
	}
	f.playStrategy = s
	return nil
}

func (f *MusicPlayerFacade) LoadPlaylist(name string) error {
	if f.playStrategy == nil {
		return errors.New("play strategy not set before loading playlist")
	}
	p, err := managers.GetPlaylistManager().GetPlaylist(name)
	if err != nil {
		return err
	}
	f.loadedPlaylist = p
	f.playStrategy.SetPlaylist(p)
	return nil
}

func (f *MusicPlayerFacade) PlaySong(song *model.Song) error {
	output, err := managers.GetDeviceManager().GetOutputDevice()
	if err != nil {
		return err
	}
	return f.audioEngine.Play(output, song)
}

func (f *MusicPlayerFacade) PauseSong(song *model.Song) error {
	if f.audioEngine.CurrentSongTitle() != song.Title {
		return fmt.Errorf("cannot pause %q; it is not currently playing", song.Title)
	}
	return f.audioEngine.Pause()
}

func (f *MusicPlayerFacade) PlayAllTracks() error {
	if f.loadedPlaylist == nil {
		return errors.New("no playlist loaded")
	}
	output, err := managers.GetDeviceManager().GetOutputDevice()
	if err != nil {
		return err
	}
	for f.playStrategy.HasNext() {
		song, err := f.playStrategy.Next()
		if err != nil {
			return err
		}
		if err := f.audioEngine.Play(output, song); err != nil {
			return err
		}
	}
	fmt.Printf("Completed playlist: %s\n", f.loadedPlaylist.Name)
	return nil
}

func (f *MusicPlayerFacade) PlayNextTrack() error {
	if f.loadedPlaylist == nil {
		return errors.New("no playlist loaded")
	}
	if !f.playStrategy.HasNext() {
		fmt.Printf("Completed playlist: %s\n", f.loadedPlaylist.Name)
		return nil
	}
	song, err := f.playStrategy.Next()
	if err != nil {
		return err
	}
	output, err := managers.GetDeviceManager().GetOutputDevice()
	if err != nil {
		return err
	}
	return f.audioEngine.Play(output, song)
}

func (f *MusicPlayerFacade) PlayPreviousTrack() error {
	if f.loadedPlaylist == nil {
		return errors.New("no playlist loaded")
	}
	if !f.playStrategy.HasPrevious() {
		fmt.Println("Already at the beginning of the playlist")
		return nil
	}
	song, err := f.playStrategy.Previous()
	if err != nil {
		return err
	}
	output, err := managers.GetDeviceManager().GetOutputDevice()
	if err != nil {
		return err
	}
	return f.audioEngine.Play(output, song)
}

func (f *MusicPlayerFacade) EnqueueNext(song *model.Song) error {
	if f.playStrategy == nil {
		return errors.New("no play strategy set")
	}
	return f.playStrategy.AddToNext(song)
}
