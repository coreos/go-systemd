package journal

import (
	"os"
	"testing"
	"time"
)

func TestJournalFollow(t *testing.T) {
	r, err := NewJournalReader(JournalReaderConfig{
		Since: time.Duration(-15) * time.Second,
		Matches: []Match{
			{
				Field: SD_JOURNAL_FIELD_SYSTEMD_UNIT,
				Value: "NetworkManager.service",
			},
		},
	})

	if err != nil {
		t.Fatalf("Error opening journal: %s", err)
	}

	if r == nil {
		t.Fatal("Got a nil reader")
	}

	defer r.Close()

	// start writing some test entries
	done := make(chan struct{}, 1)
	defer close(done)
	go func() {
		j, err := NewJournal()
		defer j.Close()

		for {
			select {
			case <-done:
				return
			default:
				if err = Printf(PriInfo, "test message %s", time.Now()); err != nil {
					t.Fatalf("Error writing to journal: %s", err)
				}

				time.Sleep(time.Second)
			}
		}
	}()

	// and follow the reader synchronously
	timeout := time.Duration(5) * time.Second
	if err = r.Follow(time.After(timeout), os.Stdout); err != ErrExpired {
		t.Fatalf("Error during follow: %s", err)
	}
}
