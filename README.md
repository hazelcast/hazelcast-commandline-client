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

# Non-interactive mode
hzc map --name myMap put --key myKey --value myValue
```

## Connecting to Hazelcast Cloud

The cluster creation and retrieving connection info can be done directly in command line using [Hazelcast Cloud CLI](https://github.com/hazelcast/hazelcast-cloud-cli).

- Authenticate to Hazelcast Cloud account:

  ```  
  hzcloud login
  -  Api Key: SAMPLE_API_KEY
  -  Api Secret: SAMPLE_API_SECRET
  ```

- Create cluster:

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
