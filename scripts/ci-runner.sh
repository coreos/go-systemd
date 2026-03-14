#!/usr/bin/env bash
set -e
set -o pipefail

PACKAGES="activation daemon dbus internal/dlopen journal login1 machine1 sdjournal unit util import1"
EXAMPLES="activation listen udpconn"

function build_source {
    go build ./...
}

function build_tests {
    rm -rf ./test_bins ; mkdir -p ./test_bins
    for pkg in ${PACKAGES}; do
        echo "  - ${pkg}"
        go test -c -o ./test_bins/${pkg}.test ./${pkg}
    done
    for ex in ${EXAMPLES}; do
        echo "  - examples/${ex}"
        go build -o ./test_bins/${ex}.example ./examples/activation/${ex}.go
    done
    # just to make sure it's buildable
    go build -o ./test_bins/journal ./examples/journal/
}

function run_in_ct {
    local image=$1
    local gover=$2
    local name="go-systemd/container-tests"
    local cidfile=/tmp/cidfile.$$
    local cid

    # Figure out Go URL, based on $gover.
    local prefix="https://go.dev/dl/" filename
    filename=$(curl -fsSL "${prefix}?mode=json&include=all" |
	jq -r --arg Ver "go$gover" '. | map(select(.version | contains($Ver))) | first | .files[] | select(.os == "linux" and .arch == "amd64" and .kind == "archive") | .filename')
    gourl="${prefix}${filename}"

    set -x
    docker pull "$image"
    docker run -i --privileged --cidfile="$cidfile" "$image" /bin/bash -e -x << EOF
if dpkg --version; then
    export DEBIAN_FRONTEND=noninteractive
    apt-get -qq update
    apt-get -qq install -y -o Dpkg::Use-Pty=0 \
	sudo build-essential curl git dbus libsystemd-dev libpam-systemd systemd-container
else # Assuming Fedora
    dnf install -qy --setopt=install_weak_deps=False --setopt=tsflags=nodocs \
	sudo curl gcc git-core dbus systemd-devel systemd-container
fi
# Fixup git.
git config --global --add safe.directory /src
# Install Go.
curl -fsSL "$gourl" | tar Cxz /usr/local
ln -s /usr/local/go/bin/go /usr/local/bin/go
go version
go env
EOF
    cid=$(cat "$cidfile")
    rm -f "$cidfile"
    docker commit "$cid" "$name"
    docker rm -f "$cid"

    echo "Starting a container with systemd..."
    docker run --shm-size=2gb -d --cidfile="$cidfile" --privileged -v "${PWD}:/src" "$name" /sbin/init --system
    cid=$(cat "$cidfile")
    rm -f "$cidfile"
    # Wait for systemd to finish, then build and run tests.
    docker exec --privileged "$cid" /bin/bash -e -c 'cd /src; ./scripts/ci-runner.sh wait_systemd build_tests run_tests'
    # Cleanup.
    docker kill "$cid"
}

function run_tests {
    pushd test_bins
    sudo -v
    for pkg in ${PACKAGES}; do
        echo "  - ${pkg}"
        sudo -E ./${pkg}.test -test.v
    done
    popd
    sudo rm -rf ./test_bins
}

function license_check {
    licRes=$(for file in $(find . -type f -iname '*.go' ! -path './vendor/*'); do
  	             head -n3 "${file}" | grep -Eq "(Copyright|generated|GENERATED)" || echo -e "  ${file}"
  	         done;)
    if [ -n "${licRes}" ]; then
        echo -e "license header checking failed:\n${licRes}"
  	    exit 253
    fi
}

while [ $# -gt 0 ]; do
    case "$1" in
    "wait_systemd" )
	shift
        echo "Waiting for systemd DBus socket..."
	while [ ! -S /run/dbus/system_bus_socket ]; do
	    sleep 0.2
	done
	echo "Waiting for systemd to finish startup..."
	# Wait for systemd to finish startup.
	systemctl is-system-running --wait || true
	;;
    "build_source" )
	shift
        echo "Building source..."
        build_source
        ;;

    "build_tests" )
	shift
        echo "Building tests..."
        build_tests
        ;;

    "run_in_ct" )
	shift
	IMAGE=$1
	shift
	GOVER=$1
	shift
	run_in_ct "$IMAGE" "$GOVER"
	;;

    "run_tests" )
	shift
        echo "Running tests..."
        run_tests
        ;;

    "license_check" )
	shift
        echo "Checking licenses..."
        license_check
        ;;

    * )
        echo "Error: unrecognized subcommand (hint: try with 'run_tests')."
        exit 1
    ;;
    esac
done
