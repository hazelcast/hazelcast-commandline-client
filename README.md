# Hazelcast CLC (Command Line Client)

## Installation

There are two ways you can install the command line client:
* Using [Brew](https://brew.sh) (**Recommended for MacOS**)
* Using script installation (Linux, MacOS)
* Currently we don't provide automated Windows installation. See [Building from source](#building-from-source).

### Installing with Brew

```
brew tap hazelcast/homebrew-hz
brew install hazelcast-commandline-client
```
**To have a superior experience, enable autocompletion on Brew:**
- For **Bash** users:
  - Execute `brew install bash-completion` and follow the printed "Caveats" section.  
    Example instruction:
    Add the following line to your ~/.bash_profile:
    ```
     [[ -r "/home/ubuntu/.linuxbrew/etc/profile.d/bash_completion.sh" ]] && . "/home/ubuntu/.linuxbrew/etc/profile.d/bash_completion.sh"
    ```
    *Note that paths may differ depending on your installation, so you should follow the Caveats section on your system.*

- For **Zsh** users
  - Follow https://docs.brew.sh/Shell-Completion#configuring-completions-in-zsh 

### Installation with script

```
curl https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/scripts/install.sh | bash
```
  
## Uninstallation

Depending on how you install the command line client, choose the uninstallation option.

### Uninstallation using Brew

```
brew uninstall hazelcast-commandline-client
brew untap hazelcast/homebrew-hz
```

### Uninstallation using script

```
~/.local/share/hz-cli/bin/uninstall.sh
```

## Usage

Make sure a Hazelcast 4 or Hazelcast 5 cluster is running.

```
# Start interactive shell
hzc

# Print help
hzc --help

# Non-interactive mode
hzc map --name myMap put --key myKey --value myValue
```
### Keyboard Shortcuts

Emacs-like keyboard shortcuts are available by default (these also are the default shortcuts in Bash shell).

| Key Binding         | Description                                     |
|---------------------|-------------------------------------------------|
| <kbd>Ctrl + A</kbd> | Go to the beginning of the line (Home)          |
| <kbd>Ctrl + E</kbd> | Go to the end of the line (End)                 |
| <kbd>Ctrl + P</kbd> | Previous command (Up arrow)                     |
| <kbd>Ctrl + N</kbd> | Next command (Down arrow)                       |
| <kbd>Ctrl + F</kbd> | Forward one character                           |
| <kbd>Ctrl + B</kbd> | Backward one character                          |
| <kbd>Ctrl + D</kbd> | Delete character under the cursor               |
| <kbd>Ctrl + H</kbd> | Delete character before the cursor (Backspace)  |
| <kbd>Ctrl + W</kbd> | Cut the word before the cursor to the clipboard |
| <kbd>Ctrl + K</kbd> | Cut the line after the cursor to the clipboard  |
| <kbd>Ctrl + U</kbd> | Cut the line before the cursor to the clipboard |
| <kbd>Ctrl + L</kbd> | Clear the screen                                |

\
With few additions:

| Key Binding          | Description                             |
|----------------------|-----------------------------------------|
| <kbd>Ctrl + C</kbd>  | Cancel running command or close the app |
| <kbd>Ctrl + -></kbd> | Go to the end of to next word           |
| <kbd>Ctrl + <-</kbd> | Go to the start of the previous word    |

## Connecting to Hazelcast Cloud

The cluster creation and retrieving connection info can be done directly in command line using [Hazelcast Cloud CLI](https://github.com/hazelcast/hazelcast-cloud-cli).

- Authenticate to Hazelcast Cloud account:

  ```  
  hzcloud login
  -  API Key: SAMPLE_API_KEY
  -  API Secret: SAMPLE_API_SECRET
  ```

- Create a cluster:

  ```
  hzcloud starter-cluster create \
  --cloud-provider=aws \
  --cluster-type=FREE \
  --name=mycluster \
  --region=us-west-2 \
  --total-memory=0.2 \
  --hazelcast-version=5.0

  > Cluster 2258 is creating. You can check the status using hzcloud starter-cluster list.
  ```
  
- Wait until the cluster is running:

  ```
  hzcloud starter-cluster list
  
  > 
  ┌────────┬────────────┬─────────┬─────────┬──────────────┬────────────────┬───────────┬─────────┐
  │ Id     │ Name       │ State   │ Version │ Memory (GiB) │ Cloud Provider │ Region    │ Is Free │
  ├────────┼────────────┼─────────┼─────────┼──────────────┼────────────────┼───────────┼─────────┤
  │ 2285   │ mycluster  │ PENDING │ 5.0     │          0.2 │ aws            │ us-west-2 │ true    │
  ├────────┼────────────┼─────────┼─────────┼──────────────┼────────────────┼───────────┼─────────┤
  │ Total: │ 1          │         │         │              │                │           │         │
  └────────┴────────────┴─────────┴─────────┴──────────────┴────────────────┴───────────┴─────────┘
  
  ...
  
  hzcloud starter-cluster list
  > 
  ┌────────┬────────────┬─────────┬─────────┬──────────────┬────────────────┬───────────┬─────────┐
  │ Id     │ Name       │ State   │ Version │ Memory (GiB) │ Cloud Provider │ Region    │ Is Free │
  ├────────┼────────────┼─────────┼─────────┼──────────────┼────────────────┼───────────┼─────────┤
  │ 2285   │ mycluster  │ RUNNING │ 5.0     │          0.2 │ aws            │ us-west-2 │ true    │
  ├────────┼────────────┼─────────┼─────────┼──────────────┼────────────────┼───────────┼─────────┤
  │ Total: │ 1          │         │         │              │                │           │         │
  └────────┴────────────┴─────────┴─────────┴──────────────┴────────────────┴───────────┴─────────┘

  ```

- Get the cluster name and discovery token:
  
  ```
  # Get cluster name
  hzcloud starter-cluster get --cluster-id 2285 --output json | jq '.releaseName'
  > "ex-1111"
  
  # Get discovery token
  hzcloud starter-cluster get --cluster-id 2285 --output json | jq '.discoveryTokens[].token'
  > "exampleHashDiscoveryToken"
  ```

- Connect to the cluster using the command line client using the credentials above:

  ```
  hzc --cluster-name <CLUSTER NAME> --cloud-token <DISCOVERY TOKEN>
  ```

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
* GNU Make (on Linux and MacOS)
* Command Prompt (on Windows) 
 
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

```
./hzc # starts the interactive mode
```
