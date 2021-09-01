# Hazelcast CLC

## Requirements

* Go 1.15 or better

## Download & Install

```
curl https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/scripts/install_raw.sh | bash
```

## Extended Download & Install

```
curl https://raw.githubusercontent.com/hazelcast/hazelcast-commandline-client/main/scripts/install_extended.sh | bash
```

## Build

### Download the Repository using Git
```
git clone https://github.com/hazelcast/hazelcast-commandline-client.git
```

### Then, Build the Project

```
cd hazelcast-commandline-client
go build -o hz-cli github.com/hazelcast/hazelcast-commandline-client
```

## Running

Make sure a Hazelcast 4 or Hazelcast 5 cluster is running.

```
# Get help
hz-cli --help
```

## Configuration
```
# Using a Default Config
# Connect to a Hazelcast Cloud cluster
# <YOUR_HAZELCAST_CLOUD_TOKEN>: token which appears on the advanced
configuration section in Hazelcast Cloud.
# <CLUSTER_NAME>: name of the cluster
hz-cli --cloud-token <YOUR_HAZELCAST_CLOUD_TOKEN> --cluster-name <CLUSTER_NAME>

# Connect to a Local Hazelcast cluster
# <ADDRESSES>: addresses of the members of the Hazelcast cluster
e.g. 192.168.1.1:5702,192.168.1.2:5703,192.168.1.3:5701
# <CLUSTER_NAME>: name of the cluster
hz-cli --address <ADDRESSES> --cluster-name <YOUR_CLUSTER_NAME>

# Using a Custom Config
# <CONFIG_PATH>: path of the target configuration
hz-cli --config <CONFIG_PATH>
```

## Operations

### Cluster Management
```
# Get state of the cluster
hz-cli cluster get-state

# Change state of the cluster
# Either of these: active | frozen | no_migration | passive
hz-cli cluster change-state --state <NEW_STATE>

# Shutdown the cluster
hz-cli cluster shutdown

# Get the version of the cluster
hz-cli cluster version
```

### Get Value & Put Value

#### Map

```
# Get from a map
hz-cli map get --name my-map --key my-key

# Put to a map
hz-cli map put --name my-map --key my-key --value my-value
```

## Examples

### Using a Default Configuration

#### Put a Value in type Map
```
hz-cli map put --name map --key a --value-type string --value "Meet"
hz-cli map get --name map --key a
> "Meet"
hz-cli map put --name map --key b --value-type json --value '{"english":"Greetings"}'
hz-cli map get --name map --key b
> {"english":"Greetings"}
```

#### Managing the Cluster
```
hz-cli cluster get-state
> {"status":"success","state":"active"}
hz-cli cluster change-state --state frozen
> {"status":"success","state":"frozen"}
hz-cli cluster shutdown
> {"status":"success"}
hz-cli cluster version
> {"status":"success","version":"5.0"}
```
