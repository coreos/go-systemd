/* Provides a low-level Go interface to the systemd journal C API.

All public methods map closely to the sd-journal API functions. See the
sd-journal.h documentation[1] for information about each function.

[1] http://www.freedesktop.org/software/systemd/man/sd-journal.html
*/
package journal

/*
#cgo pkg-config: libsystemd-journal
#include <systemd/sd-journal.h>
#include <stdlib.h>
#include <syslog.h>

int go_sd_journal_print(int priority, char* s) {
	int r;
	r = sd_journal_print(priority, "%s", s);
	return r;
}
*/
import "C"
import (
	"fmt"
	"sync"
	"time"
	"unsafe"
)

// Journal entry field strings which correspond to:
// http://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html
const (
	SD_JOURNAL_FIELD_SYSTEMD_UNIT = "_SYSTEMD_UNIT"
	SD_JOURNAL_FIELD_MESSAGE      = "MESSAGE"
	SD_JOURNAL_FIELD_PID          = "_PID"
	SD_JOURNAL_FIELD_UID          = "_UID"
	SD_JOURNAL_FIELD_GID          = "_GID"
	SD_JOURNAL_FIELD_HOSTNAME     = "_HOSTNAME"
	SD_JOURNAL_FIELD_MACHINE_ID   = "_MACHINE_ID"
)

// Journal event constants
const (
	SD_JOURNAL_NOP        = int(C.SD_JOURNAL_NOP)
	SD_JOURNAL_APPEND     = int(C.SD_JOURNAL_APPEND)
	SD_JOURNAL_INVALIDATE = int(C.SD_JOURNAL_INVALIDATE)
)

// A Journal is a Go wrapper of an sd_journal structure.
type Journal struct {
	cjournal *C.sd_journal
	mu       sync.Mutex
}

// A Match is a convenience wrapper to describe filters supplied to AddMatch.
type Match struct {
	Field string
	Value string
}

// String returns a string representation of a Match suitable for use with AddMatch.
func (m *Match) String() string {
	return m.Field + "=" + m.Value
}

// NewJournal returns a new Journal instance pointing to the local journal
func NewJournal() (*Journal, error) {
	j := &Journal{}
	err := C.sd_journal_open(&j.cjournal, C.SD_JOURNAL_LOCAL_ONLY)

	if err < 0 {
		return nil, fmt.Errorf("failed to open journal: %s", err)
	}

	return j, nil
}

func (j *Journal) Close() error {
	j.mu.Lock()
	C.sd_journal_close(j.cjournal)
	j.mu.Unlock()
	return nil
}

func (j *Journal) AddMatch(match string) error {
	m := C.CString(match)
	defer C.free(unsafe.Pointer(m))

	j.mu.Lock()
	C.sd_journal_add_match(j.cjournal, unsafe.Pointer(m), C.size_t(len(match)))
	j.mu.Unlock()

	return nil
}

func (j *Journal) Next() (int, error) {
	j.mu.Lock()
	r := C.sd_journal_next(j.cjournal)
	j.mu.Unlock()

	if r < 0 {
		return int(r), fmt.Errorf("failed to iterate journal: %d", r)
	}

	return int(r), nil
}

func (j *Journal) Previous() (uint64, error) {
	j.mu.Lock()
	r := C.sd_journal_previous(j.cjournal)
	j.mu.Unlock()

	if r < 0 {
		return uint64(r), fmt.Errorf("failed to iterate journal: %d", r)
	}

	return uint64(r), nil
}

func (j *Journal) PreviousSkip(skip uint64) (uint64, error) {
	j.mu.Lock()
	r := C.sd_journal_previous_skip(j.cjournal, C.uint64_t(skip))
	j.mu.Unlock()

	if r < 0 {
		return uint64(r), fmt.Errorf("failed to iterate journal: %d", r)
	}

	return uint64(r), nil
}

func (j *Journal) GetData(field string) (string, error) {
	f := C.CString(field)
	defer C.free(unsafe.Pointer(f))

	var d unsafe.Pointer
	var l C.size_t

	j.mu.Lock()
	err := C.sd_journal_get_data(j.cjournal, f, &d, &l)
	j.mu.Unlock()

	if err < 0 {
		return "", fmt.Errorf("failed to read message: %d", err)
	}

	msg := C.GoStringN((*C.char)(d), C.int(l))
	return msg, nil
}

func (j *Journal) GetRealtimeUsec() (uint64, error) {
	var usec C.uint64_t

	j.mu.Lock()
	r := C.sd_journal_get_realtime_usec(j.cjournal, &usec)
	j.mu.Unlock()

	if r < 0 {
		return 0, fmt.Errorf("error getting timestamp for entry: %d", r)
	}

	return uint64(usec), nil
}

func (j *Journal) Print(priority int, message string) error {
	m := C.CString(message)
	defer C.free(unsafe.Pointer(m))

	err := C.go_sd_journal_print(C.LOG_INFO, m)

	if err != 0 {
		return fmt.Errorf("failed to print message: %s", err)
	}

	return nil
}

func (j *Journal) SeekTail() error {
	j.mu.Lock()
	err := C.sd_journal_seek_tail(j.cjournal)
	j.mu.Unlock()

	if err != 0 {
		return fmt.Errorf("failed to seek to tail of journal: %s", err)
	}

	return nil
}

func (j *Journal) SeekRealtimeUsec(usec uint64) error {
	j.mu.Lock()
	err := C.sd_journal_seek_realtime_usec(j.cjournal, C.uint64_t(usec))
	j.mu.Unlock()

	if err != 0 {
		return fmt.Errorf("failed to seek to %d: %d", usec, int(err))
	}

	return nil
}

func (j *Journal) Wait(timeout time.Duration) int {
	to := uint64(time.Now().Add(timeout).Unix() / 1000)
	j.mu.Lock()
	r := C.sd_journal_wait(j.cjournal, C.uint64_t(to))
	j.mu.Unlock()

	return int(r)
}
