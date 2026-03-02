package main

import (
	enums "app/enum"
	"app/player"
	"fmt"
	//"enums"
)

func main() {
	player := player.GetMusicPlayerApplication()

	// Populate library
	player.CreateSongInLibrary("Kesariya", "Arijit Singh", "/music/kesariya.mp3")
	player.CreateSongInLibrary("Chaiyya Chaiyya", "Sukhwinder Singh", "/music/chaiyya_chaiyya.mp3")
	player.CreateSongInLibrary("Tum Hi Ho", "Arijit Singh", "/music/tum_hi_ho.mp3")
	player.CreateSongInLibrary("Jai Ho", "A. R. Rahman", "/music/jai_ho.mp3")
	player.CreateSongInLibrary("Zinda", "Siddharth Mahadevan", "/music/zinda.mp3")

	// Create playlist and add songs
	must(player.CreatePlaylist("Bollywood Vibes"))
	must(player.AddSongToPlaylist("Bollywood Vibes", "Kesariya"))
	must(player.AddSongToPlaylist("Bollywood Vibes", "Chaiyya Chaiyya"))
	must(player.AddSongToPlaylist("Bollywood Vibes", "Tum Hi Ho"))
	must(player.AddSongToPlaylist("Bollywood Vibes", "Jai Ho"))

	// Connect device
	player.ConnectAudioDevice(enums.Bluetooth)

	// Play/pause a single song
	must(player.PlaySingleSong("Zinda"))
	must(player.PauseCurrentSong("Zinda"))
	must(player.PlaySingleSong("Zinda")) // resume

	fmt.Println("\n-- Sequential Playback --\n")
	must(player.SelectPlayStrategy(enums.Sequential))
	must(player.LoadPlaylist("Bollywood Vibes"))
	must(player.PlayAllTracksInPlaylist())

	fmt.Println("\n-- Random Playback --\n")
	must(player.SelectPlayStrategy(enums.Random))
	must(player.LoadPlaylist("Bollywood Vibes"))
	must(player.PlayAllTracksInPlaylist())

	fmt.Println("\n-- Custom Queue Playback --\n")
	must(player.SelectPlayStrategy(enums.CustomQueue))
	must(player.LoadPlaylist("Bollywood Vibes"))
	must(player.QueueSongNext("Kesariya"))
	must(player.QueueSongNext("Tum Hi Ho"))
	must(player.PlayAllTracksInPlaylist())

	fmt.Println("\n-- Play Previous in Sequential --\n")
	must(player.SelectPlayStrategy(enums.Sequential))
	must(player.LoadPlaylist("Bollywood Vibes"))
	must(player.PlayAllTracksInPlaylist())
	must(player.PlayPreviousTrackInPlaylist())
	must(player.PlayPreviousTrackInPlaylist())
}

func must(err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
