// +build !linux

package sdjournal

import "C"
import (
	"errors"
	"time"
	"unsafe"
)

// Journal entry field strings which correspond to:
// http://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html
const (
	// User Journal Fields
	SD_JOURNAL_FIELD_MESSAGE           = "MESSAGE"
	SD_JOURNAL_FIELD_MESSAGE_ID        = "MESSAGE_ID"
	SD_JOURNAL_FIELD_PRIORITY          = "PRIORITY"
	SD_JOURNAL_FIELD_CODE_FILE         = "CODE_FILE"
	SD_JOURNAL_FIELD_CODE_LINE         = "CODE_LINE"
	SD_JOURNAL_FIELD_CODE_FUNC         = "CODE_FUNC"
	SD_JOURNAL_FIELD_ERRNO             = "ERRNO"
	SD_JOURNAL_FIELD_SYSLOG_FACILITY   = "SYSLOG_FACILITY"
	SD_JOURNAL_FIELD_SYSLOG_IDENTIFIER = "SYSLOG_IDENTIFIER"
	SD_JOURNAL_FIELD_SYSLOG_PID        = "SYSLOG_PID"

	// Trusted Journal Fields
	SD_JOURNAL_FIELD_PID                       = "_PID"
	SD_JOURNAL_FIELD_UID                       = "_UID"
	SD_JOURNAL_FIELD_GID                       = "_GID"
	SD_JOURNAL_FIELD_COMM                      = "_COMM"
	SD_JOURNAL_FIELD_EXE                       = "_EXE"
	SD_JOURNAL_FIELD_CMDLINE                   = "_CMDLINE"
	SD_JOURNAL_FIELD_CAP_EFFECTIVE             = "_CAP_EFFECTIVE"
	SD_JOURNAL_FIELD_AUDIT_SESSION             = "_AUDIT_SESSION"
	SD_JOURNAL_FIELD_AUDIT_LOGINUID            = "_AUDIT_LOGINUID"
	SD_JOURNAL_FIELD_SYSTEMD_CGROUP            = "_SYSTEMD_CGROUP"
	SD_JOURNAL_FIELD_SYSTEMD_SESSION           = "_SYSTEMD_SESSION"
	SD_JOURNAL_FIELD_SYSTEMD_UNIT              = "_SYSTEMD_UNIT"
	SD_JOURNAL_FIELD_SYSTEMD_USER_UNIT         = "_SYSTEMD_USER_UNIT"
	SD_JOURNAL_FIELD_SYSTEMD_OWNER_UID         = "_SYSTEMD_OWNER_UID"
	SD_JOURNAL_FIELD_SYSTEMD_SLICE             = "_SYSTEMD_SLICE"
	SD_JOURNAL_FIELD_SELINUX_CONTEXT           = "_SELINUX_CONTEXT"
	SD_JOURNAL_FIELD_SOURCE_REALTIME_TIMESTAMP = "_SOURCE_REALTIME_TIMESTAMP"
	SD_JOURNAL_FIELD_BOOT_ID                   = "_BOOT_ID"
	SD_JOURNAL_FIELD_MACHINE_ID                = "_MACHINE_ID"
	SD_JOURNAL_FIELD_HOSTNAME                  = "_HOSTNAME"
	SD_JOURNAL_FIELD_TRANSPORT                 = "_TRANSPORT"

	// Address Fields
	SD_JOURNAL_FIELD_CURSOR              = "__CURSOR"
	SD_JOURNAL_FIELD_REALTIME_TIMESTAMP  = "__REALTIME_TIMESTAMP"
	SD_JOURNAL_FIELD_MONOTONIC_TIMESTAMP = "__MONOTONIC_TIMESTAMP"
)

// Journal event constants
const (
	SD_JOURNAL_NOP        = int(C.SD_JOURNAL_NOP)
	SD_JOURNAL_APPEND     = int(C.SD_JOURNAL_APPEND)
	SD_JOURNAL_INVALIDATE = int(C.SD_JOURNAL_INVALIDATE)
)

const (
	// IndefiniteWait is a sentinel value that can be passed to
	// sdjournal.Wait() to signal an indefinite wait for new journal
	// events. It is implemented as the maximum value for a time.Duration:
	// https://github.com/golang/go/blob/e4dcf5c8c22d98ac9eac7b9b226596229624cb1d/src/time/time.go#L434
	IndefiniteWait time.Duration = 1<<63 - 1
)

var (
	// ErrNoTestCursor gets returned when using TestCursor function and cursor
	// parameter is not the same as the current cursor position.
	ErrNoTestCursor = errors.New("Cursor parameter is not the same as current position")
)

// Journal is a Go wrapper of an sd_journal structure.
type Journal struct{}

// JournalEntry represents all fields of a journal entry plus address fields.
type JournalEntry struct {
	Fields             map[string]string
	Cursor             string
	RealtimeTimestamp  uint64
	MonotonicTimestamp uint64
}

// Match is a convenience wrapper to describe filters supplied to AddMatch.
type Match struct {
	Field string
	Value string
}

// String returns a string representation of a Match suitable for use with AddMatch.
func (m *Match) String() string {
	panic("not implemented")
}

// NewJournal returns a new Journal instance pointing to the local journal
func NewJournal() (j *Journal, err error) {
	panic("not implemented")
}

// NewJournalFromDir returns a new Journal instance pointing to a journal residing
// in a given directory.
func NewJournalFromDir(path string) (j *Journal, err error) {
	panic("not implemented")
}

// NewJournalFromFiles returns a new Journal instance pointing to a journals residing
// in a given files.
func NewJournalFromFiles(paths ...string) (j *Journal, err error) {
	panic("not implemented")
}

// Close closes a journal opened with NewJournal.
func (j *Journal) Close() error {
	panic("not implemented")
}

// AddMatch adds a match by which to filter the entries of the journal.
func (j *Journal) AddMatch(match string) error {
	panic("not implemented")
}

// AddDisjunction inserts a logical OR in the match list.
func (j *Journal) AddDisjunction() error {
	panic("not implemented")
}

// AddConjunction inserts a logical AND in the match list.
func (j *Journal) AddConjunction() error {
	panic("not implemented")
}

// FlushMatches flushes all matches, disjunctions and conjunctions.
func (j *Journal) FlushMatches() {
	panic("not implemented")
}

// Next advances the read pointer into the journal by one entry.
func (j *Journal) Next() (uint64, error) {
	panic("not implemented")
}

// NextSkip advances the read pointer by multiple entries at once,
// as specified by the skip parameter.
func (j *Journal) NextSkip(skip uint64) (uint64, error) {
	panic("not implemented")
}

// Previous sets the read pointer into the journal back by one entry.
func (j *Journal) Previous() (uint64, error) {
	panic("not implemented")
}

// PreviousSkip sets back the read pointer by multiple entries at once,
// as specified by the skip parameter.
func (j *Journal) PreviousSkip(skip uint64) (uint64, error) {
	panic("not implemented")
}

func (j *Journal) getData(field string) (unsafe.Pointer, C.int, error) {
	panic("not implemented")
}

// GetData gets the data object associated with a specific field from the
// the journal entry referenced by the last completed Next/Previous function
// call. To call GetData, you must have first called one of these functions.
func (j *Journal) GetData(field string) (string, error) {
	panic("not implemented")
}

// GetDataValue gets the data object associated with a specific field from the
// journal entry referenced by the last completed Next/Previous function call,
// returning only the value of the object. To call GetDataValue, you must first
// have called one of the Next/Previous functions.
func (j *Journal) GetDataValue(field string) (string, error) {
	panic("not implemented")
}

// GetDataBytes gets the data object associated with a specific field from the
// journal entry referenced by the last completed Next/Previous function call.
// To call GetDataBytes, you must first have called one of these functions.
func (j *Journal) GetDataBytes(field string) ([]byte, error) {
	panic("not implemented")
}

// GetDataValueBytes gets the data object associated with a specific field from the
// journal entry referenced by the last completed Next/Previous function call,
// returning only the value of the object. To call GetDataValueBytes, you must first
// have called one of the Next/Previous functions.
func (j *Journal) GetDataValueBytes(field string) ([]byte, error) {
	panic("not implemented")
}

// GetEntry returns a full representation of the journal entry referenced by the
// last completed Next/Previous function call, with all key-value pairs of data
// as well as address fields (cursor, realtime timestamp and monotonic timestamp).
// To call GetEntry, you must first have called one of the Next/Previous functions.
func (j *Journal) GetEntry() (*JournalEntry, error) {
	panic("not implemented")
}

// SetDataThreshold sets the data field size threshold for data returned by
// GetData. To retrieve the complete data fields this threshold should be
// turned off by setting it to 0, so that the library always returns the
// complete data objects.
func (j *Journal) SetDataThreshold(threshold uint64) error {
	panic("not implemented")
}

// GetRealtimeUsec gets the realtime (wallclock) timestamp of the journal
// entry referenced by the last completed Next/Previous function call. To
// call GetRealtimeUsec, you must first have called one of the Next/Previous
// functions.
func (j *Journal) GetRealtimeUsec() (uint64, error) {
	panic("not implemented")
}

// GetMonotonicUsec gets the monotonic timestamp of the journal entry
// referenced by the last completed Next/Previous function call. To call
// GetMonotonicUsec, you must first have called one of the Next/Previous
// functions.
func (j *Journal) GetMonotonicUsec() (uint64, error) {
	panic("not implemented")
}

// GetCursor gets the cursor of the last journal entry reeferenced by the
// last completed Next/Previous function call. To call GetCursor, you must
// first have called one of the Next/Previous functions.
func (j *Journal) GetCursor() (string, error) {
	panic("not implemented")
}

// TestCursor checks whether the current position in the journal matches the
// specified cursor
func (j *Journal) TestCursor(cursor string) error {
	panic("not implemented")
}

// SeekHead seeks to the beginning of the journal, i.e. the oldest available
// entry. This call must be followed by a call to Next before any call to
// Get* will return data about the first element.
func (j *Journal) SeekHead() error {
	panic("not implemented")
}

// SeekTail may be used to seek to the end of the journal, i.e. the most recent
// available entry. This call must be followed by a call to Next before any
// call to Get* will return data about the last element.
func (j *Journal) SeekTail() error {
	panic("not implemented")
}

// SeekRealtimeUsec seeks to the entry with the specified realtime (wallclock)
// timestamp, i.e. CLOCK_REALTIME. This call must be followed by a call to
// Next/Previous before any call to Get* will return data about the sought entry.
func (j *Journal) SeekRealtimeUsec(usec uint64) error {
	panic("not implemented")
}

// SeekCursor seeks to a concrete journal cursor. This call must be
// followed by a call to Next/Previous before any call to Get* will return
// data about the sought entry.
func (j *Journal) SeekCursor(cursor string) error {
	panic("not implemented")
}

// Wait will synchronously wait until the journal gets changed. The maximum time
// this call sleeps may be controlled with the timeout parameter.  If
// sdjournal.IndefiniteWait is passed as the timeout parameter, Wait will
// wait indefinitely for a journal change.
func (j *Journal) Wait(timeout time.Duration) int {
	panic("not implemented")
}

// GetUsage returns the journal disk space usage, in bytes.
func (j *Journal) GetUsage() (uint64, error) {
	panic("not implemented")
}

// GetUniqueValues returns all unique values for a given field.
func (j *Journal) GetUniqueValues(field string) ([]string, error) {
	panic("not implemented")
}

// GetCatalog retrieves a message catalog entry for the journal entry referenced
// by the last completed Next/Previous function call. To call GetCatalog, you
// must first have called one of these functions.
func (j *Journal) GetCatalog() (string, error) {
	panic("not implemented")
}

// Process will process events for the journal.
func (j *Journal) Process() error {
	panic("not implemented")
}

// RestartData resets the data enumeration index to the beginning of the entry.
func (j *Journal) RestartData() error {
	panic("not implemented")
}
