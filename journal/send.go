package journal

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type Priority int

const (
	PRI_EMERG Priority = iota
	PRI_ALERT
	PRI_CRIT
	PRI_ERR
	PRI_WARNING
	PRI_NOTICE
	PRI_INFO
	PRI_DEBUG
)

var conn net.Conn

func init() {
	var err error
	conn, err = net.Dial("unixgram", "/run/systemd/journal/socket")
	if err != nil {
		fmt.Fprintln(os.Stderr, "socket connection error:", err)
	}
}

func Enabled() bool {
	return conn != nil
}

func Send(message string, priority Priority, vars map[string]string) {
	if conn == nil {
		fmt.Fprintln(os.Stderr, "journal error: connection uninitialized")
		return
	}

	data := new(bytes.Buffer)
	appendVariable(data, "PRIORITY", strconv.Itoa(int(priority)))
	appendVariable(data, "MESSAGE", message)
	for k, v := range vars {
		appendVariable(data, k, v)
	}

	_, err := io.Copy(conn, data)
	if err != nil && isSocketSpaceError(err) {
		file, err := tempFd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "journal error:", err)
			return
		}
		_, err = io.Copy(file, data)
		if err != nil {
			fmt.Fprintln(os.Stderr, "journal error:", err)
			return
		}

		rights := syscall.UnixRights(int(file.Fd()))

		/* this connection should always be a UnixConn, but better safe than sorry */
		unixConn, ok := conn.(*net.UnixConn)
		if !ok {
			fmt.Fprintln(os.Stderr, "journal error: can't send file through non-Unix connection")
			return
		}

		unixConn.WriteMsgUnix([]byte{}, rights, nil)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, "journal error:", err)
	}
}

func appendVariable(w io.Writer, name, value string) {
	if !validVarName(name) {
		fmt.Fprintln(os.Stderr, "journal error: variable name contains invalid character, ignoring")
		return
	}
	if strings.ContainsRune(value, '\n') {
		/* When the value contains a newline, we write:
		 * - the variable name, followed by a newline
		 * - the size (in 64bit little endian format)
		 * - the data, followed by a newline
		 */
		fmt.Fprintln(w, name)
		w.Write(toBytes(uint64(len(value))))
		fmt.Fprintln(w, value)
	} else {
		/* just write the variable and value all on one line */
		fmt.Fprintf(w, "%s=%s\n", name, value)
	}
}

func validVarName(name string) bool {
	/* The variable name must be in uppercase and consist only of characters,
	 * numbers and underscores, and may not begin with an underscore. (from the docs)
	 */
	valid := true
	valid = valid && name[0] != '_'
	for _, c := range name {
		valid = valid && ('A' <= c && c <= 'Z') || ('0' <= c && c <='9') || c == '_'
	}
	return valid
}

func isSocketSpaceError(err error) bool {
	opErr, ok := err.(*net.OpError)
	if !ok {
		return false
	}

	sysErr, ok := opErr.Err.(syscall.Errno)
	if !ok {
		return false
	}

	return sysErr == syscall.EMSGSIZE || sysErr == syscall.ENOBUFS
}

func toBytes(n uint64) []byte {
	var b [8]byte
	for i := 0; i < 8; i++ {
		b[i] = byte(n >> uint(i*8))
	}
	return b[:]
}

func tempFd() (*os.File, error) {
	file, err := ioutil.TempFile("/dev/shm/", "journal.XXXXX")
	if err != nil {
		return nil, err
	}
	syscall.Unlink(file.Name())
	if err != nil {
		return nil, err
	}
	return file, nil
}
