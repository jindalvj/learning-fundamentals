package external

import "fmt"

type BluetoothSpeakerAPI struct{}

func (b *BluetoothSpeakerAPI) PlaySoundViaBluetooth(data string) {
	fmt.Printf("[BluetoothSpeaker] Playing: %s\n", data)
}

type WiredSpeakerAPI struct{}

func (w *WiredSpeakerAPI) PlaySoundViaCable(data string) {
	fmt.Printf("[WiredSpeaker] Playing: %s\n", data)
}

type HeadphonesAPI struct{}

func (h *HeadphonesAPI) PlaySoundViaJack(data string) {
	fmt.Printf("[Headphones] Playing: %s\n", data)
}
