# go-systemd

Go bindings to systemd socket activation, journal and D-BUS APIs.

## Socket Activation

See an example in `examples/activation/httpserver.go`. For easy debugging use
`/usr/lib/systemd/systemd-activate`

## D-Bus

The D-Bus API lets you start, stop and introspect systemd units. The API docs are here:

http://godoc.org/github.com/coreos/go-systemd/dbus

### Debugging

Create `/etc/dbus-1/system-local.conf` that looks like this:

```
<!DOCTYPE busconfig PUBLIC
"-//freedesktop//DTD D-Bus Bus Configuration 1.0//EN"
"http://www.freedesktop.org/standards/dbus/1.0/busconfig.dtd">
<busconfig>
    <policy user="root">
        <allow eavesdrop="true"/>
        <allow eavesdrop="true" send_destination="*"/>
    </policy>
</busconfig>
```
