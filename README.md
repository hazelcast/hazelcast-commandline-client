# Hazelcast CLC

## Installation

There are three ways you can install CLC:

* Using [Brew](https://brew.sh) (Linux, MacOS)
* Using script installation (Linux, MacOS)
* Install wizard (Windows only)

### Installing with Brew

```
$ brew tap hazelcast/homebrew-hz
$ brew install hazelcast-clc
```

**To have a superior experience, enable autocompletion on Brew:**

* **Bash** users:
  * Execute `brew install bash-completion` and follow the printed "Caveats" section.  
    Example instruction:
    Add the following line to your ~/.bash_profile:
    ```
     [[ -r "/home/ubuntu/.linuxbrew/etc/profile.d/bash_completion.sh" ]] && . "/home/ubuntu/.linuxbrew/etc/profile.d/bash_completion.sh"
    ```
    *Note that paths may differ depending on your installation, so you should follow the Caveats section on your system.*

* **Zsh** users
  * Follow https://docs.brew.sh/Shell-Completion#configuring-completions-in-zsh 

### Installation with script

If you are using Linux or MacOS, you can run the following Bash script to install 

```
$ curl https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/scripts/install.sh | bash
```

### Installing using the Windows Install Wizard

If you are using a recent version of Windows, you may prefer to install CLC using the Install Wizard we provide.
You can download the Install Wizard at: https://github.com/hazelcast/hazelcast-commandline-client/releases
  
## Uninstallation

Depending on how you install the command line client, choose the uninstallation option.

### Uninstallation using Brew

```
$ brew uninstall hazelcast-clc
$ brew untap hazelcast/homebrew-hz
```

### Uninstallation using script

```
$ bash ~/.local/share/clc/bin/uninstall.sh
```

### Uninstallation using the Windows Install Wizard

If you have installed CLC using the Windows Install Wizard, you can use the Settings/Apps menu to uninstall it.

## Usage

Make sure a Hazelcast 4 or Hazelcast 5 cluster is running.

```
# Start interactive shell
$ clc

# Print help
$ clc --help

# Non-interactive mode
$ clc map put myKey myValue
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

## Connecting to Hazelcast Cloud




## SSL Configuration

You can use the following configuration file to enable SSL support:
```
ssl:
    enabled: true
    servername: "HOSTNAME-FOR-SERVER"
    # or: insecureskipverify: true
hazelcast:
  cluster:
    security:
      credentials:
        username: "OPTIONAL USERNAME"
        password: "OPTIONAL PASSWORD"
    name: "CLUSTER-NAME"
    network:
      addresses:
        - "localhost:5701"
```

Mutual authentication is also supported:
```
ssl:
    enabled: true
    servername: "HOSTNAME-FOR-SERVER"
    # insecureskipverify: true
    capath: "/tmp/ca.pem"
    certpath: "/tmp/cert.pem"
    keypath: "/tmp/key.pem"
    keypassword: "PASSWORD FOR THE KEY"
hazelcast:
  cluster:
    security:
      credentials:
        username: "OPTIONAL USERNAME"
        password: "OPTIONAL PASSWORD"
    name: "CLUSTER-NAME"
    network:
      addresses:
        - "localhost:5701"
```

Cloud SSL configuration:
```
ssl:
    enabled: true
    capath: "/tmp/ca.pem"
    certpath: "/tmp/cert.pem"
    keypath: "/tmp/key.pem"
    keypassword: "PASSWORD FOR THE KEY"
hazelcast:
  cluster:
    name: "CLUSTER NAME"
    cloud:
      token: "HAZELCAST CLOUD TOKEN"
      enabled: true
```

## More examples

```
# Get from a map
hzc map get --name my-map --key my-key

# Put to a map
hzc map put --name my-map --key my-key --value my-value

# Get state of the cluster
hzc cluster get-state

# Work with JSON values
hzc map put --name map --key b --value-type json --value '{"english":"Greetings"}'
hzc map get --name map --key b
> {"english":"Greetings"}

# Change state of the cluster
# Either of these: active | frozen | no_migration | passive
hzc cluster change-state --state <NEW_STATE>

# Shutdown the cluster
hzc cluster shutdown

# Get the version of the cluster
hzc cluster version
```

## Configuration
```
# Using Custom Config
# <CONFIG_PATH>: path of the target configuration
hzc --config <CONFIG_PATH>

# Connect to a Local Hazelcast cluster
# <ADDRESSES>: addresses of the members of the Hazelcast cluster
e.g. 192.168.1.1:5702,192.168.1.2:5703,192.168.1.3:5701
# <CLUSTER_NAME>: name of the cluster
hzc --address <ADDRESSES> --cluster-name <YOUR_CLUSTER_NAME>
```

## Building from source

The following targets are tested and supported.
The prior versions of the given targets would also work, but that's not tested. 

* Ubuntu 18.04 or better.
* MacOS 12 or better.
* Windows 10 or better.

### Requirements

* Go 1.18 or better
* Git
* GNU Make (on Linux and MacOS)
* Command Prompt or Powershell (on Windows) 
* go-winres: https://github.com/tc-hib/go-winres (on Windows)
 
### 1. Download the source

You can acquire the source using Git:

```
git clone https://github.com/hazelcast/hazelcast-commandline-client.git
```

Or download the source archive and extract it:

```
https://github.com/hazelcast/hazelcast-commandline-client/archive/refs/heads/main.zip
```

### 2. Build the project

```
cd hazelcast-commandline-client
make
```

### Finally, run the project

CLC starts the interactive mode by default.

On Linux and MacOS:
```
./hzc
```

On Windows:
```
hzc.exe
```
