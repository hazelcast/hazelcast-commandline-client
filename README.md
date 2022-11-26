# Hazelcast CLC

## Usage Summary

### Non-interactive Mode

Run commands:
```
$ clc map put my-key my-value
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
CLC> \map put my-key my-value
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
| <kbd>Ctrl + C</kbd> | Cancel running command or close the app        |
| <kbd>Ctrl + -></kbd>| Go to the end of to next word                  |
| <kbd>Ctrl + <-</kbd>| Go to the start of the previous word           |

## Connecting to Viridian Serverless

1. If you don't have a running Viridian Serverless cluster, follow the steps in [Step 1. Start a Viridian Serverless Development Cluster](https://docs.hazelcast.com/cloud/get-started#step-1-start-a-viridian-serverless-development-cluster) to create a cluster.
  Both development and production clusters will work very well.
2. Download the Go client sample for your cluster from the Viridian Console. The sample is typically a Zip file with the following name format: "hazelcast-cloud-go-sample-client-CLUSTER-ID-default.zip". For instance: `hazelcast-cloud-csharp-sample-client-pr-3814-default.zip` 
3. Import the configuration with CLC:
  ```
  $ clc config import ~/hazelcast-cloud-go-sample-client-pr-3814-default.zip
  ```
4. The configuration will be imported with the name as the cluster ID. Check that the configuration is known to CLC:
  ```
  $ clc config list
  default
  pr-3814
  ```
5. In order to use this configuration, use `-c CONFIG_NAME` flag whenever you run CLC:
  ```
  $ clc -c pr-3814 map put my-key my-value
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

* Ubuntu 18.04 or better.
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

CLC starts the interactive mode by default.

On Linux and MacOS:
```
./build/clc
```

On Windows:
```
.\build\clc.exe
```
