package journal

import (
	"fmt"
	"log/syslog"
	"os"
	"testing"
	"time"
)

func TestJournalFollow(t *testing.T) {
	r, err := NewJournalReader(JournalReaderConfig{
		Since: time.Duration(-15) * time.Second,
		Matches: []Match{
			Match{
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
	go func() {
		j, err := NewJournal()
		defer j.Close()

	writer:
		for {
			select {
			case <-done:
				break writer
			default:
				err = j.Print(int(syslog.LOG_INFO), fmt.Sprintf("test message %s", time.Now()))

				if err != nil {
					t.Fatalf("Error writing to journal: %s\n", err)
				}

				time.Sleep(time.Duration(1) * time.Second)
			}
		}
	}()

	// and follow the reader synchronously
	timeout := time.Duration(5) * time.Second
	err = r.Follow(time.After(timeout), os.Stdout)

	// shut down the test writer
	close(done)

	if err != nil {
		t.Fatalf("Error during follow: %s", err)
	}
}
