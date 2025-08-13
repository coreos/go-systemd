// Package unit provides utilities for parsing, serializing, and manipulating
// systemd unit files. It supports both reading unit file content into Go data
// structures and writing Go data structures back to unit file format.
//
// The package provides functionality to:
//   - Parse systemd unit files into [UnitOption] and [UnitSection] structures
//   - Serialize Go structures back into unit file format
//   - Escape and unescape unit names according to systemd conventions
//
// Unit files are configuration files that describe how systemd should manage
// services, sockets, devices, and other system resources.
package unit
