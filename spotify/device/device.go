package device

import (
	"app/external"
	"app/model"
	"fmt"
)

// IAudioOutputDevice is the target interface all adapters implement
type IAudioOutputDevice interface {
	PlayAudio(song *model.Song)
}

// --- Bluetooth Adapter ---
type BluetoothSpeakerAdapter struct {
	api *external.BluetoothSpeakerAPI
}

func NewBluetoothSpeakerAdapter(api *external.BluetoothSpeakerAPI) *BluetoothSpeakerAdapter {
	return &BluetoothSpeakerAdapter{api: api}
}

func (b *BluetoothSpeakerAdapter) PlayAudio(song *model.Song) {
	payload := fmt.Sprintf("%s by %s", song.Title, song.Artist)
	b.api.PlaySoundViaBluetooth(payload)
}

// --- Wired Speaker Adapter ---
type WiredSpeakerAdapter struct {
	api *external.WiredSpeakerAPI
}

func NewWiredSpeakerAdapter(api *external.WiredSpeakerAPI) *WiredSpeakerAdapter {
	return &WiredSpeakerAdapter{api: api}
}

func (w *WiredSpeakerAdapter) PlayAudio(song *model.Song) {
	payload := fmt.Sprintf("%s by %s", song.Title, song.Artist)
	w.api.PlaySoundViaCable(payload)
}

// --- Headphones Adapter ---
type HeadphonesAdapter struct {
	api *external.HeadphonesAPI
}

func NewHeadphonesAdapter(api *external.HeadphonesAPI) *HeadphonesAdapter {
	return &HeadphonesAdapter{api: api}
}

func (h *HeadphonesAdapter) PlayAudio(song *model.Song) {
	payload := fmt.Sprintf("%s by %s", song.Title, song.Artist)
	h.api.PlaySoundViaJack(payload)
}
