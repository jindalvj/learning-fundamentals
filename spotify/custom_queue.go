package strategies

import (
	"errors"
	"musicplayer/models"
)

// CustomQueueStrategy plays from a priority queue first, then falls back to sequential
type CustomQueueStrategy struct {
	playlist     *models.Playlist
	currentIndex int
	nextQueue    []*models.Song // songs to play next (FIFO)
	prevStack    []*models.Song // history for going back (LIFO)
}

func NewCustomQueueStrategy() *CustomQueueStrategy {
	return &CustomQueueStrategy{currentIndex: -1}
}

func (c *CustomQueueStrategy) SetPlaylist(playlist *models.Playlist) {
	c.playlist = playlist
	c.currentIndex = -1
	c.nextQueue = nil
	c.prevStack = nil
}

func (c *CustomQueueStrategy) HasNext() bool {
	if c.playlist == nil {
		return false
	}
	return len(c.nextQueue) > 0 || c.currentIndex+1 < c.playlist.Size()
}

func (c *CustomQueueStrategy) Next() (*models.Song, error) {
	if c.playlist == nil || c.playlist.Size() == 0 {
		return nil, errors.New("no playlist loaded or playlist is empty")
	}

	var song *models.Song

	if len(c.nextQueue) > 0 {
		// Dequeue from priority queue
		song = c.nextQueue[0]
		c.nextQueue = c.nextQueue[1:]

		// Sync currentIndex to where this song sits in the playlist
		for i, s := range c.playlist.Songs() {
			if s == song {
				c.currentIndex = i
				break
			}
		}
	} else {
		// Sequential fallback
		if c.currentIndex+1 >= c.playlist.Size() {
			return nil, errors.New("no more songs in playlist")
		}
		c.currentIndex++
		song = c.playlist.Songs()[c.currentIndex]
	}

	c.prevStack = append(c.prevStack, song)
	return song, nil
}

func (c *CustomQueueStrategy) HasPrevious() bool {
	return len(c.prevStack) > 0
}

func (c *CustomQueueStrategy) Previous() (*models.Song, error) {
	if c.playlist == nil || c.playlist.Size() == 0 {
		return nil, errors.New("no playlist loaded or playlist is empty")
	}
	if len(c.prevStack) == 0 {
		return nil, errors.New("no previous song available")
	}

	last := len(c.prevStack) - 1
	song := c.prevStack[last]
	c.prevStack = c.prevStack[:last]

	// Sync index
	for i, s := range c.playlist.Songs() {
		if s == song {
			c.currentIndex = i
			break
		}
	}
	return song, nil
}

func (c *CustomQueueStrategy) AddToNext(song *models.Song) error {
	if song == nil {
		return errors.New("cannot enqueue nil song")
	}
	c.nextQueue = append(c.nextQueue, song)
	return nil
}
