# Hazelcast CLC

## Installation
There are two ways you can install command line client:
* With Brew([Homebrew Package Manager](https://brew.sh)) [**Recommended**]
* With custom installation

| Installation Method 	| Install                                                                                   	| Uninstall 	|
|---------------------	|-------------------------------------------------------------------------------------------	|-----------	|
| Brew  | <pre>brew tap utku-caglayan/hazelcast-clc<br>brew install hazelcast-commandline-client</pre> 	|    <pre>brew uninstall hazelcast-commandline-client<br>brew untap utku-caglayan/hazelcast-clc</pre>       	|
| Custom Script       	| `curl https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/scripts/install.sh \| bash`                                                                                          	|`~/.local/share/hz-cli/bin/uninstall.sh`|

:bangbang: Please follow the instructions to enable autocompletion for Brew
### **To have superior experience, enable autocompletion on Brew:**
- If you are using brew on **bash**: <br>`brew install bash-completion` and follow the printed "Caveats" section.<br>Example instruction:<br>Add the following line to your ~/.bash_profile: `[[ -r "/home/ubuntu/.linuxbrew/etc/profile.d/bash_completion.sh" ]] && . "/home/ubuntu/.linuxbrew/etc/profile.d/bash_completion.sh"`.<br>Note that paths may differ depending on your installation, so you should follow the Caveats section on your system.

- If you are using brew on **zsh**:    
  https://docs.brew.sh/Shell-Completion#configuring-completions-in-zsh

## Build
### Requirements
* Go 1.15 or better
### Download the Repository using Git
```
git clone https://github.com/hazelcast/hazelcast-commandline-client.git
```

### Then, Build the Project

```
cd hazelcast-commandline-client
go build -o hzc github.com/hazelcast/hazelcast-commandline-client
```

## Usage

Make sure a Hazelcast 4 or Hazelcast 5 cluster is running.

```
# Get help
hzc --help
# or interactively
hzc
```

## Configuration
```
# Using a Default Config
# Connect to a Hazelcast Cloud cluster
# <YOUR_HAZELCAST_CLOUD_TOKEN>: token which appears on the advanced
configuration section in Hazelcast Cloud.
# <CLUSTER_NAME>: name of the cluster
hzc --cloud-token <YOUR_HAZELCAST_CLOUD_TOKEN> --cluster-name <CLUSTER_NAME>

# Connect to a Local Hazelcast cluster
# <ADDRESSES>: addresses of the members of the Hazelcast cluster
e.g. 192.168.1.1:5702,192.168.1.2:5703,192.168.1.3:5701
# <CLUSTER_NAME>: name of the cluster
hzc --address <ADDRESSES> --cluster-name <YOUR_CLUSTER_NAME>

# Using a Custom Config
# <CONFIG_PATH>: path of the target configuration
hzc --config <CONFIG_PATH>
```

## Operations

### Cluster Management
```
# Get state of the cluster
hzc cluster get-state

# Change state of the cluster
# Either of these: active | frozen | no_migration | passive
hzc cluster change-state --state <NEW_STATE>

# Shutdown the cluster
hzc cluster shutdown

# Get the version of the cluster
hzc cluster version
```

### Get Value & Put Value

#### Map

```
# Get from a map
hzc map get --name my-map --key my-key

# Put to a map
hzc map put --name my-map --key my-key --value my-value
```

## Examples

### Using a Default Configuration

#### Put a Value in type Map
```
hzc map put --name map --key a --value-type string --value "Meet"
hzc map get --name map --key a
> "Meet"
hzc map put --name map --key b --value-type json --value '{"english":"Greetings"}'
hzc map get --name map --key b
> {"english":"Greetings"}
```

#### Managing the Cluster
```
hzc cluster get-state
> {"status":"success","state":"active"}
hzc cluster change-state --state frozen
> {"status":"success","state":"frozen"}
hzc cluster shutdown
> {"status":"success"}
hzc cluster version
> {"status":"success","version":"5.0"}
```