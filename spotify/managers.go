package managers

import (
	"errors"
	"fmt"
	"musicplayer/device"
	"musicplayer/enums"
	"musicplayer/factory"
	"musicplayer/models"
	"musicplayer/strategies"
	"sync"
)

// ─── DeviceManager ───────────────────────────────────────────────────────────

type DeviceManager struct {
	currentDevice device.IAudioOutputDevice
}

var (
	deviceManagerOnce     sync.Once
	deviceManagerInstance *DeviceManager
)

func GetDeviceManager() *DeviceManager {
	deviceManagerOnce.Do(func() {
		deviceManagerInstance = &DeviceManager{}
	})
	return deviceManagerInstance
}

func (d *DeviceManager) Connect(deviceType enums.DeviceType) {
	d.currentDevice = factory.CreateDevice(deviceType)
	fmt.Printf("%s device connected\n", deviceType)
}

func (d *DeviceManager) GetOutputDevice() (device.IAudioOutputDevice, error) {
	if d.currentDevice == nil {
		return nil, errors.New("no output device connected")
	}
	return d.currentDevice, nil
}

func (d *DeviceManager) HasOutputDevice() bool {
	return d.currentDevice != nil
}

// ─── PlaylistManager ─────────────────────────────────────────────────────────

type PlaylistManager struct {
	playlists map[string]*models.Playlist
}

var (
	playlistManagerOnce     sync.Once
	playlistManagerInstance *PlaylistManager
)

func GetPlaylistManager() *PlaylistManager {
	playlistManagerOnce.Do(func() {
		playlistManagerInstance = &PlaylistManager{
			playlists: make(map[string]*models.Playlist),
		}
	})
	return playlistManagerInstance
}

func (pm *PlaylistManager) CreatePlaylist(name string) error {
	if _, exists := pm.playlists[name]; exists {
		return fmt.Errorf("playlist %q already exists", name)
	}
	pm.playlists[name] = models.NewPlaylist(name)
	return nil
}

func (pm *PlaylistManager) AddSongToPlaylist(playlistName string, song *models.Song) error {
	p, exists := pm.playlists[playlistName]
	if !exists {
		return fmt.Errorf("playlist %q not found", playlistName)
	}
	return p.AddSong(song)
}

func (pm *PlaylistManager) GetPlaylist(name string) (*models.Playlist, error) {
	p, exists := pm.playlists[name]
	if !exists {
		return nil, fmt.Errorf("playlist %q not found", name)
	}
	return p, nil
}

// ─── StrategyManager ─────────────────────────────────────────────────────────

type StrategyManager struct {
	sequential  *strategies.SequentialPlayStrategy
	random      *strategies.RandomPlayStrategy
	customQueue *strategies.CustomQueueStrategy
}

var (
	strategyManagerOnce     sync.Once
	strategyManagerInstance *StrategyManager
)

func GetStrategyManager() *StrategyManager {
	strategyManagerOnce.Do(func() {
		strategyManagerInstance = &StrategyManager{
			sequential:  strategies.NewSequentialPlayStrategy(),
			random:      strategies.NewRandomPlayStrategy(),
			customQueue: strategies.NewCustomQueueStrategy(),
		}
	})
	return strategyManagerInstance
}

func (sm *StrategyManager) GetStrategy(strategyType enums.PlayStrategyType) (strategies.PlayStrategy, error) {
	switch strategyType {
	case enums.Sequential:
		return sm.sequential, nil
	case enums.Random:
		return sm.random, nil
	case enums.CustomQueue:
		return sm.customQueue, nil
	default:
		return nil, errors.New("unknown strategy type")
	}
}
