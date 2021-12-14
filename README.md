# Hazelcast CLC

## Installation

There are two ways you can install the command line client:
* Using [Brew](https://brew.sh) [**Recommended**]
* Using script installation

### Installing with Brew [Recommended]

```
brew tap utku-caglayan/hazelcast-clc
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
brew untap utku-caglayan/hazelcast-clc
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
```

### More examples

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
# Using a Default Config
# Connect to a Hazelcast Cloud cluster
# <YOUR_HAZELCAST_CLOUD_TOKEN>: token which appears on the advanced configuration section in Hazelcast Cloud.
# <CLUSTER_NAME>: name of the cluster
hzc --cloud-token <YOUR_HAZELCAST_CLOUD_TOKEN> --cluster-name <CLUSTER_NAME>

# Connect to a Local Hazelcast cluster
# <ADDRESSES>: addresses of the members of the Hazelcast cluster
e.g. 192.168.1.1:5702,192.168.1.2:5703,192.168.1.3:5701
# <CLUSTER_NAME>: name of the cluster
hzc --address <ADDRESSES> --cluster-name <YOUR_CLUSTER_NAME>

# Using Custom Config
# <CONFIG_PATH>: path of the target configuration
hzc --config <CONFIG_PATH>
```

## Build from source

### Requirements
* Go 1.15 or better
 
### Download the repository using Git
```
git clone https://github.com/hazelcast/hazelcast-commandline-client.git
```

### Then, build the project

```
cd hazelcast-commandline-client
go build -o hzc github.com/hazelcast/hazelcast-commandline-client
```
