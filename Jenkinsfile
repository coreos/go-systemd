matrixJob('Periodic go-systemd builder from dsl') {
    label('master')
    displayName('Periodic go-systemd builder (master branch) from dsl')

    scm {
        git ('https://github.com/coreos/go-systemd.git', '*/master')
    }

    concurrentBuild()

    triggers {
        cron('@daily')
    }

    axes {
        label('os_type', 'debian-testing', 'fedora-24', 'fedora-25')
    }

    wrappers {
        buildNameSetter {
            template('go-systemd master (periodic #${BUILD_NUMBER})')
            runAtStart(true)
            runAtEnd(true)
        }
        timeout {
            absolute(25)
        }
    }

    steps {
        shell('./scripts/jenkins/go-systemd-master.sh')
    }
}