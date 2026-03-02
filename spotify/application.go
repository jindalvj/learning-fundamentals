package app

import (
	"fmt"
	"musicplayer/enums"
	"musicplayer/managers"
	"musicplayer/models"
	"sync"
)

type MusicPlayerApplication struct {
	songLibrary []*models.Song
}

var (
	appOnce     sync.Once
	appInstance *MusicPlayerApplication
)

func GetMusicPlayerApplication() *MusicPlayerApplication {
	appOnce.Do(func() {
		appInstance = &MusicPlayerApplication{}
	})
	return appInstance
}

func (a *MusicPlayerApplication) CreateSongInLibrary(title, artist, path string) {
	a.songLibrary = append(a.songLibrary, &models.Song{
		Title:    title,
		Artist:   artist,
		FilePath: path,
	})
}

func (a *MusicPlayerApplication) FindSongByTitle(title string) (*models.Song, error) {
	for _, s := range a.songLibrary {
		if s.Title == title {
			return s, nil
		}
	}
	return nil, fmt.Errorf("song %q not found in library", title)
}

func (a *MusicPlayerApplication) CreatePlaylist(name string) error {
	return managers.GetPlaylistManager().CreatePlaylist(name)
}

func (a *MusicPlayerApplication) AddSongToPlaylist(playlistName, songTitle string) error {
	song, err := a.FindSongByTitle(songTitle)
	if err != nil {
		return err
	}
	return managers.GetPlaylistManager().AddSongToPlaylist(playlistName, song)
}

func (a *MusicPlayerApplication) ConnectAudioDevice(deviceType enums.DeviceType) {
	GetMusicPlayerFacade().ConnectDevice(deviceType)
}

func (a *MusicPlayerApplication) SelectPlayStrategy(strategyType enums.PlayStrategyType) error {
	return GetMusicPlayerFacade().SetPlayStrategy(strategyType)
}

func (a *MusicPlayerApplication) LoadPlaylist(name string) error {
	return GetMusicPlayerFacade().LoadPlaylist(name)
}

func (a *MusicPlayerApplication) PlaySingleSong(title string) error {
	song, err := a.FindSongByTitle(title)
	if err != nil {
		return err
	}
	return GetMusicPlayerFacade().PlaySong(song)
}

func (a *MusicPlayerApplication) PauseCurrentSong(title string) error {
	song, err := a.FindSongByTitle(title)
	if err != nil {
		return err
	}
	return GetMusicPlayerFacade().PauseSong(song)
}

func (a *MusicPlayerApplication) PlayAllTracksInPlaylist() error {
	return GetMusicPlayerFacade().PlayAllTracks()
}

func (a *MusicPlayerApplication) PlayPreviousTrackInPlaylist() error {
	return GetMusicPlayerFacade().PlayPreviousTrack()
}

func (a *MusicPlayerApplication) QueueSongNext(title string) error {
	song, err := a.FindSongByTitle(title)
	if err != nil {
		return err
	}
	return GetMusicPlayerFacade().EnqueueNext(song)
}
