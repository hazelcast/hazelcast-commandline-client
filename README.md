# Hazelcast CLC

## Installation

We provide binaries for the popular platforms at our [Releases](https://github.com/hazelcast/hazelcast-commandline-client/releases) page.
In order to install CLC:

* Download the release package for your platform,
* Extract it,
* Optionally move the `clc` binary to somewhere in your *PATH*, so it can be run in any terminal without additional settings.

Currently we provide precompiled binaries of CLC for the following platforms and architectures:

* Linux/amd64
* Windows/amd64
* MacOS/amd64
* MacOS/arm64

Additionally, we provide an installer for Windows 10 and up.
The installer can install CLC for either system-wide or just for the user.
It adds the `clc` binary automatically to the `$PATH`, so it can be run in any terminal without additional settings.

If your platform is not one of the above, you may want to compile CLC yourself. Our build process is very simple and doesn't have many dependencies.
In most cases just running `make` is sufficient to build CLC if you have the latest [Go](https://go.dev/) compiler installed.
See [Building from source](#building-from-source) section.

## Usage Summary

### Home Directory

CLC stores all configuration, logs and other files in its home directory.
We will refer to that directory as `$CLC_HOME`.
You can use `clc home` command to show where is `$CLC_HOME`:
```
$ clc home
/home/guest/.local/share/clc
```

### Configuration

CLC has a simple YAML configuration, usually named `config.yaml`.
This file can exist anywhere in the file system, and can be used with the `--config` (or `-c`) flag:

```
$ clc -c test/config.yaml
```

If there is a `config.yaml` in the same directory with the CLC binary and the configuration was not explicitly set, CLC tries to load that configuration file:
```
$ ls -lh
total 17M
-rwxrwxr-x 1 yuce yuce  17M Nov 26 23:11 clc*
-rw------- 1 yuce yuce  200 Nov 26 23:12 config.yaml
```

`configs` directory in `$CLC_HOME` is special, it contains all the configurations known to CLC.
Known configurations can be directly specified by their names, instead of the full path.
`clc config list` command lists the configurations known to CLC:
```
# List configurations
$ clc config list
default
pr-3066

# Start CLC shell with configuration named pr-3066 
$ clc -c pr-3066
```

If no configuration is specified, the `default` configuration is used if it exists.

#### Configuration format

All paths in the configuration are relative to the parent directory of the configuration file.

* cluster
  * name: Name of the cluster. By default `dev`.
  * address: Address of a member in the cluster. By default `localhost:5701`.
  * discovery-token: {hazelcast-cloud} discovery token.

* ssl
  * ca-path: TLS CA certificate path.
  * cert-path: TLS certificate path.
  * key-path: TLS mutual authentication key certificate path.
  * key-password: Password for the key certificate.

* log
  * path: Path to the log file, or `stderr`. By default, the logs are written to `$CLC_HOME/logs` with the current date as the name.
  * level: Log level, one of: `debug`, `info`, `warn`, `error`. The default is `info`.

Here's a sample {hazelcast-cloud} Serverless configuration:
```
cluster:
  name: "pr-3814"  
  discovery-token: "HY8eR7X...."
ssl:
  ca-path: "ca.pem"
  cert-path: "cert.pem"
  key-path: "key.pem"
  key-password: "a6..."
```

### Logging

CLC doesn't log to the screen by default, in order to reduce clutter.
By default, logs are saved in `$CLC_HOME/logs`, creating a new log file per day.
In order to log to a different file, or to stderr (usually the screen), use the `--log.path` flag:

```
# log to object-list.log
$ clc object list --log.path object-list.log

# log to screen
$ clc object list --log.path stderr
```

By default, logs with level `info` and above are logged.
You can use the `--log.level` flag to change the level.
Supported levels: `debug`, `info`, `warn`, `error`

```
# log only errors
$ clc object list --log.level error
```


### Non-interactive Mode

Run commands:
```
$ clc map set my-key my-value
```

Get help:
```
$ clc --help
```

### Interactive (Shell) Mode

Start interactive shell:
```
$ clc
```

Run SQL commands:
```
CLC> select * from cities;
```

Run CLC commands:
```
CLC> \map set my-key my-value
```
 
### Keyboard Shortcuts

The following keyboard shortcuts are available in the interactive-mode:

| Key Binding         | Description                                    |
|---------------------|------------------------------------------------|
| <kbd>Ctrl + A</kbd> | Go to the beginning of the line (Home)         |
| <kbd>Ctrl + E</kbd> | Go to the end of the line (End)                |
| <kbd>Ctrl + P</kbd> | Previous command (Up arrow)                    |
| <kbd>Ctrl + N</kbd> | Next command (Down arrow)                      |
| <kbd>Ctrl + F</kbd> | Forward one character                          |
| <kbd>Ctrl + B</kbd> | Backward one character                         |
| <kbd>Ctrl + D</kbd> | Delete character under the cursor              |
| <kbd>Ctrl + H</kbd> | Delete character before the cursor (Backspace) |
| <kbd>Ctrl + W</kbd> | Cut the word before the cursor                 |
| <kbd>Ctrl + K</kbd> | Cut the line after the cursor                  |
| <kbd>Ctrl + U</kbd> | Cut the line before the cursor                 |
| <kbd>Ctrl + L</kbd> | Clear the screen                               |
| <kbd>Ctrl + C</kbd> | Cancel running command        |

## Connecting to Viridian Serverless

1. If you don't have a running Viridian Serverless cluster, follow the steps in [Step 1. Start a Viridian Serverless Development Cluster](https://docs.hazelcast.com/cloud/get-started#step-1-start-a-viridian-serverless-development-cluster) to create a cluster.
  Both development and production clusters will work very well.
2. Download the Go client sample for your cluster from the Viridian Console. The sample is typically a Zip file with the following name format: "hazelcast-cloud-go-sample-client-CLUSTER-ID-default.zip". For instance: `hazelcast-cloud-csharp-sample-client-pr-3814-default.zip` 
3. Import the configuration with CLC with `my-cluster` as the name:
  ```
  $ clc config import my-cluster ~/hazelcast-cloud-go-sample-client-pr-3814-default.zip
  ```
4. Check that the configuration is known to CLC:
  ```
  $ clc config list
  default
  my-cluster
  ```
5. In order to use this configuration, use `-c CONFIG_NAME` flag whenever you run CLC:
  ```
  $ clc -c my-cluster map set my-key my-value
  ```

## Generating auto-completion

CLC supports auto-completion in the non-interactive mode for the following shells:
* Bash
* Fish
* Powershell
* Zsh

Run `clc completion SHELL-NAME` command to generate an autocompletion file.
The instructions for each shell is different and can be access by running `clc completion SHELL-NAME --help`:
```
# show the instructions to enable auto-completion for bash
$ clc completion bash --help

# generate the auto-completion for bash and save it to clc.bash
$ clc completion bash > clc.bash
```

## Building from source

The following targets are tested and supported.
The prior versions of the given targets would also work, but that's not tested. 

* Ubuntu 22.04 or better.
* MacOS 12 or better.
* Windows 10 or better.

### Requirements

* Go 1.19 or better
* Git
* GNU Make (on Linux and MacOS)
* Command Prompt or Powershell (on Windows) 
* go-winres: https://github.com/tc-hib/go-winres (on Windows)
 
### 1. Download the source

You can acquire the source using Git:

```
$ git clone https://github.com/hazelcast/hazelcast-commandline-client.git
```

Or download the source archive from https://github.com/hazelcast/hazelcast-commandline-client/archive/refs/heads/main.zip and extract it.

### 2. Build the project

```
$ cd hazelcast-commandline-client
$ make
```

The `clc` or `clc.exe` binary is created in the `build` directory.

### Finally, run the project

CLC starts the in interactive mode by default.

On Linux and MacOS:
```
./build/clc
```

On Windows:
```
.\build\clc.exe
```
