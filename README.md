# Hazelcast CLC

Hazelcast CLC is a command line tool that enables developers to use Hazelcast data structures, run SQL queries, scaffold projects, manage Viridian clusters and more.

Apart from this README, this repository contains:
* [Examples](https://github.com/hazelcast/hazelcast-commandline-client/blob/main/examples) to get you started with various features of CLC.
* Some useful [Advanced Scripts](https://github.com/hazelcast/hazelcast-commandline-client/blob/main/scripts).

## Documentation

Our documentation is hosted at: https://docs.hazelcast.com/clc/latest/overview.

## Installation

### Linux / macOS

We provide an easy to use install script for the following platforms:
* Linux/amd64
* Linux/arm64
* Linux/arm
* macOS/amd64 (Intel)
* macOS/arm64 (M1, M2, ...)

You can install the latest version of CLC using:

  curl https://hazelcast.com/clc/install.sh | bash

### Windows

You can download an installer for amd64 / Intel from https://github.com/hazelcast/hazelcast-commandline-client/releases/latest.
The installer is named `hazelcast-clc-setup_VERSION_amd64.exe`

The installer can install CLC either system-wide or just for the user.
It adds the `clc` binary automatically to the `$PATH`, so it can be run in any terminal without additional settings.

### Manual Download

You can download CLC manually from our [Releases](https://github.com/hazelcast/hazelcast-commandline-client/releases) page.
The latest GA release is always at: https://github.com/hazelcast/hazelcast-commandline-client/releases/latest.

On macOS, binaries downloaded outside of AppStore require your intervention to run.
The install script automatically handles this, but if you downloaded a release package you can do it manually:
```
$ xattr -d com.apple.quarantine CLC_FOLDER/clc
```
Use the correct path instead of `CLC_FOLDER` in the command above.

### Building from Source

The latest open source release of CLC is v5.3.3 and it can be cloned from:
https://github.com/hazelcast/hazelcast-commandline-client/tree/v5.3.3

CLC v5.3.4 and up are binary only releases.
You can find all our relases at:
https://github.com/hazelcast/hazelcast-commandline-client/releases.

## Help

* Join our Slack channel: https://hazelcastcommunity.slack.com/channels/clc. You can get an invite at: https://slack.hazelcast.com/
* Check out this link for our Enterprise support: https://hazelcast.com/services/support/

## Other Resources

* [Introducing CLC: The New Hazelcast Command-Line Experience](https://hazelcast.com/blog/introducing-clc-the-new-hazelcast-command-line-experience/) [Blog article]
* [The New Hazelcast Command-Line Experience](https://www.youtube.com/watch?v=lIj7jEV-jp4) [Video]
* [Creating and Managing Real-Time Kafka Pipelines with Hazelcast CLC](https://www.youtube.com/watch?v=Q_9Y9yQBzIY) [Video]
