package downloadqueue

import (
	"aura/database"
	"sync"
)

// QueueEvent is broadcast to all subscribed SSE clients whenever the
// download queue changes.
type QueueEvent struct {
	Type     string                      `json:"type"` // "snapshot" | "job_added" | "job_started" | "job_progress" | "job_finished"
	Job      *database.DownloadQueueJob  `json:"job,omitempty"`
	Jobs     []database.DownloadQueueJob `json:"jobs,omitempty"`
	Progress *QueueProgress              `json:"progress,omitempty"`
}

// QueueProgress is the image currently being downloaded for a job.
type QueueProgress struct {
	JobID          int64  `json:"job_id"`
	MediaItemTitle string `json:"media_item_title"`
	ImageType      string `json:"image_type"`
	SeasonNumber   *int   `json:"season_number,omitempty"`
	EpisodeNumber  *int   `json:"episode_number,omitempty"`
}

type broadcaster struct {
	mu   sync.Mutex
	subs map[chan QueueEvent]struct{}
}

var QueueBroadcaster = &broadcaster{subs: map[chan QueueEvent]struct{}{}}

// Subscribe registers a new listener. The returned channel is buffered;
// slow consumers have events dropped rather than blocking publishers.
func (b *broadcaster) Subscribe() chan QueueEvent {
	ch := make(chan QueueEvent, 16)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *broadcaster) Unsubscribe(ch chan QueueEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.subs[ch]; ok {
		delete(b.subs, ch)
		close(ch)
	}
}

func (b *broadcaster) Publish(evt QueueEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- evt:
		default:
			// Slow subscriber: drop the event, the client will get true
			// state from the next mutation or a fresh snapshot on reconnect.
		}
	}
}
