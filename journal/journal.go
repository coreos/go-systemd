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
	"errors"
	"fmt"
	"unsafe"
)

// Journal entry field strings which correspond to:
// http://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html
const (
	SD_JOURNAL_FIELD_SYSTEMD_UNIT = "_SYSTEMD_UNIT"
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

func NewJournal() (*Journal, error) {
	j := &Journal{}
	err := C.sd_journal_open(&j.cjournal, C.SD_JOURNAL_LOCAL_ONLY)

	if err < 0 {
		return nil, errors.New(fmt.Sprintf("Failed to open journal: %s\n", err))
	}

	return j, nil
}

func (j *Journal) Close() error {
	C.sd_journal_close(j.cjournal)
	return nil
}

func (j *Journal) AddMatch(match string) error {
	m := C.CString(match)
	defer C.free(unsafe.Pointer(m))

	C.sd_journal_add_match(j.cjournal, unsafe.Pointer(m), C.size_t(len(match)))
	return nil
}

func (j *Journal) Next() (int, error) {
	r := C.sd_journal_next(j.cjournal)

	if r < 0 {
		return int(r), errors.New(fmt.Sprintf("Failed to iterate journal: %d\n", r))
	}

	return int(r), nil
}

func (j *Journal) Previous() (uint64, error) {
	r := C.sd_journal_previous(j.cjournal)

	if r < 0 {
		return uint64(r), errors.New(fmt.Sprintf("Failed to iterate journal: %d\n", r))
	}

	return uint64(r), nil
}

func (j *Journal) PreviousSkip(skip uint64) (uint64, error) {
	r := C.sd_journal_previous_skip(j.cjournal, C.uint64_t(skip))

	if r < 0 {
		return uint64(r), errors.New(fmt.Sprintf("Failed to iterate journal: %d\n", r))
	}

	return uint64(r), nil
}

func (j *Journal) GetData(field string) (string, error) {
	f := C.CString(field)
	defer C.free(unsafe.Pointer(f))

	var d unsafe.Pointer
	var l C.size_t

	err := C.sd_journal_get_data(j.cjournal, f, &d, &l)

	if err < 0 {
		return "", errors.New(fmt.Sprintf("Failed to read message: %d\n", err))
	} else {
		msg := C.GoStringN((*C.char)(d), C.int(l))
		return msg, nil
	}
}

func (j *Journal) GetRealtimeUsec() (uint64, error) {
	var usec C.uint64_t

	r := C.sd_journal_get_realtime_usec(j.cjournal, &usec)

	if r < 0 {
		return 0, errors.New(fmt.Sprintf("Error getting timestamp for entry: %d\n", r))
	}

	return uint64(usec), nil
}

func (j *Journal) Print(priority int, message string) error {
	m := C.CString(message)
	defer C.free(unsafe.Pointer(m))

	err := C.go_sd_journal_print(C.LOG_INFO, m)

	if err == 0 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("Failed to print message: %s\n", err))
	}
}

func (j *Journal) SeekTail() error {
	err := C.sd_journal_seek_tail(j.cjournal)

	if err != 0 {
		return errors.New(fmt.Sprintf("Failed to seek to tail of journal: %s\n", err))
	} else {
		return nil
	}
}

func (j *Journal) SeekRealtimeUsec(usec uint64) error {
	err := C.sd_journal_seek_realtime_usec(j.cjournal, C.uint64_t(usec))

	if err != 0 {
		return errors.New(fmt.Sprintf("Failed to seek to %d: %d\n", usec, int(err)))
	}

	return nil
}

func (j *Journal) Wait(timeout uint64) int {
	r := C.sd_journal_wait(j.cjournal, C.uint64_t(timeout))

	return int(r)
}
