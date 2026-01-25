// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package activation

import (
	"crypto/tls"
	"net"
)

// Listeners returns a slice containing a net.Listener for each matching socket type
// passed to this process.
//
// The order of the file descriptors is preserved in the returned slice.
// Nil values are used to fill any gaps. For example if systemd were to return file descriptors
// corresponding with "udp, tcp, tcp", then the slice would contain {nil, net.Listener, net.Listener}
func Listeners() ([]net.Listener, error) {
	return ListenersWithOptions()
}

// ListenersWithNames maps a listener name to a set of net.Listener instances.
func ListenersWithNames() (map[string][]net.Listener, error) {
	return ListenersWithNamesWithOptions()
}

// TLSListeners returns a slice containing a net.listener for each matching TCP socket type
// passed to this process.
// It uses default Listeners func and forces TCP sockets handlers to use TLS based on tlsConfig.
func TLSListeners(tlsConfig *tls.Config) ([]net.Listener, error) {
	return TLSListenersWithOptions(tlsConfig)
}

// TLSListenersWithNames maps a listener name to a net.Listener with
// the associated TLS configuration.
func TLSListenersWithNames(tlsConfig *tls.Config) (map[string][]net.Listener, error) {
	return TLSListenersWithNamesWithOptions(tlsConfig)
}

// TLSListenersWithOptions returns a slice containing a net.Listener for each matching TCP socket type
// passed to this process.
// It uses ListenersWithOptions func and forces TCP sockets handlers to use TLS based on tlsConfig.
//
//   - The socket activation file descriptors are consumed and closed by default. This
//     can be changed by passing in the [UseMethod] option. See [Method] for more details.
//   - Unsetting the corresponding environment variables is the default behavior. This
//     can be changed by passing in the [UnsetEnv] option.
func TLSListenersWithOptions(tlsConfig *tls.Config, opts ...option) ([]net.Listener, error) {
	listeners, err := ListenersWithOptions(opts...)

	if listeners == nil || err != nil {
		return nil, err
	}

	if tlsConfig != nil {
		for i, l := range listeners {
			// Activate TLS only for TCP sockets
			if l.Addr().Network() == "tcp" {
				listeners[i] = tls.NewListener(l, tlsConfig)
			}
		}
	}

	return listeners, err
}

// TLSListenersWithNamesWithOptions maps a listener name to a net.Listener with
// the associated TLS configuration.
//
//   - The socket activation file descriptors are consumed and closed by default. This
//     can be changed by passing in the [UseMethod] option. See [Method] for more details.
//   - Unsetting the corresponding environment variables is the default behavior. This
//     can be changed by passing in the [UnsetEnv] option.
func TLSListenersWithNamesWithOptions(tlsConfig *tls.Config, opts ...option) (map[string][]net.Listener, error) {
	listeners, err := ListenersWithNamesWithOptions(opts...)

	if listeners == nil || err != nil {
		return nil, err
	}

	if tlsConfig != nil {
		for _, ll := range listeners {
			// Activate TLS only for TCP sockets
			for i, l := range ll {
				if l.Addr().Network() == "tcp" {
					ll[i] = tls.NewListener(l, tlsConfig)
				}
			}
		}
	}

	return listeners, err
}

// ListenersWithOptions returns a slice containing a net.Listener for each matching socket type
// passed to this process.
//
//   - The socket activation file descriptors are consumed and closed by default. This
//     can be changed by passing in the [UseMethod] option. See [Method] for more details.
//   - Unsetting the corresponding environment variables is the default behavior. This
//     can be changed by passing in the [UnsetEnv] option.
func ListenersWithOptions(opts ...option) ([]net.Listener, error) {
	o := options{unsetEnv: true}
	o.apply(opts...)

	files := Files(o.unsetEnv)
	listeners := make([]net.Listener, len(files))

	for i, f := range files {
		if pc, err := net.FileListener(f); err == nil {
			listeners[i] = pc

			o.method.Apply(f)
		}
	}

	return listeners, nil
}

// ListenersWithNamesWithOptions maps a listener name to a set of net.Listener instances.
//
//   - The socket activation file descriptors are consumed and closed by default. This
//     can be changed by passing in the [UseMethod] option. See [Method] for more details.
//   - Unsetting the corresponding environment variables is the default behavior. This
//     can be changed by passing in the [UnsetEnv] option.
func ListenersWithNamesWithOptions(opts ...option) (map[string][]net.Listener, error) {
	o := options{unsetEnv: true}
	o.apply(opts...)

	files := Files(o.unsetEnv)
	listeners := map[string][]net.Listener{}

	for _, f := range files {
		if pc, err := net.FileListener(f); err == nil {
			listeners[f.Name()] = append(listeners[f.Name()], pc)

			o.method.Apply(f)
		}
	}

	return listeners, nil
}
