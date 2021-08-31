# Hazelcast CLC

## Requirements

* Go 1.15 or better

## Download & Install

```
curl https://github.com/hazelcast/hazelcast-commandline-client/blob/main/install.sh | bash
```

## Build

### Download the Repository from Git CLI
```
git clone https://github.com/hazelcast/hazelcast-commandline-client.git
```

### Download the Repository from GitHub CLI
```
gh repo clone hazelcast/hazelcast-commandline-client
```

### Then, Build the Project

```
cd hazelcast-commandline-client
go build -o hz-cli github.com/hazelcast/hazelcast-commandline-client
```

## Running

Make sure a Hazelcast v5.2021.07.1 instance is running.

```
# Get help
hz-cli --help

# Put to a map
hz-cli map put --name my-map --key my-key --value my-value

# Get from a map
hz-cli map get --name my-map --key my-key
```

## Examples

### Default Configuration
#### Register a Value in type Map
```
hz-cli map put --name my-map --key a --value-type string --value "Meet"
hz-cli map get --name my-map --key a
> "Meet"
hz-cli map put --name my-map --key b --value-type json --value '{"english":"Greetings"}'
hz-cli map get --name my-map --key b
> {"english":"Greetings"}
```

#### Manage the Cluster
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

### Custom Configuration via Command Line

#### Operate on Hazelcast Cloud

##### Register a Value in type Map
```
hz-cli --cloud-token <YOUR_HAZELCAST_CLOUD_TOKEN> --cluster-name <YOUR_CLUSTER_NAME> map put --name map --key a --value-type string --value "Meet"
hz-cli --cloud-token <YOUR_HAZELCAST_CLOUD_TOKEN> --cluster-name <YOUR_CLUSTER_NAME> map get --name map --key a
> "Meet"
hz-cli --cloud-token <YOUR_HAZELCAST_CLOUD_TOKEN> --cluster-name <YOUR_CLUSTER_NAME> map put --name map --key b --value-type json --value '{"english":"Greetings"}'
hz-cli --cloud-token <YOUR_HAZELCAST_CLOUD_TOKEN> --cluster-name <YOUR_CLUSTER_NAME> map get --name map --key b
> {"english":"Greetings"}
```

##### Manage the Cluster
*Cluster management operations are not permitted in Hazelcast Cloud*

#### Operate on Local Hazelcast instance

##### Register a Value in type Map
```
hz-cli --address 192.168.1.1:5701,192.168.1.1:5702 --cluster-name <YOUR_CLUSTER_NAME> map put --name my-map --key a --value-type string --value "Meet"
hz-cli --address 192.168.1.1:5701,192.168.1.1:5702 --cluster-name <YOUR_CLUSTER_NAME> map get --name my-map --key a
> "Meet"
hz-cli --address 192.168.1.1:5701,192.168.1.1:5702 --cluster-name <YOUR_CLUSTER_NAME> map put --name my-map --key b --value-type json --value '{"english":"Greetings"}'
hz-cli --address 192.168.1.1:5701,192.168.1.1:5702 --cluster-name <YOUR_CLUSTER_NAME> map get --name my-map --key b
> {"english":"Greetings"}
```

##### Manage the Cluster
```
hz-cli --address 192.168.1.1:5701,192.168.1.1:5702 --cluster-name <YOUR_CLUSTER_NAME> cluster get-state
> {"status":"success","state":"active"}
hz-cli --address 192.168.1.1:5701,192.168.1.1:5702 --cluster-name <YOUR_CLUSTER_NAME> cluster change-state --state frozen
> {"status":"success","state":"frozen"}
hz-cli --address 192.168.1.1:5701,192.168.1.1:5702 --cluster-name <YOUR_CLUSTER_NAME> cluster shutdown
> {"status":"success"}
hz-cli --address 192.168.1.1:5701,192.168.1.1:5702 --cluster-name <YOUR_CLUSTER_NAME> cluster version
> {"status":"success","version":"5.0"}
```

#### Use custom configuration file

##### Register a Value in type Map
```
hz-cli --config <CONFIG_PATH> map put --name my-map --key a --value-type string --value "Meet"
hz-cli --config <CONFIG_PATH> map get --name my-map --key a
> "Meet"
hz-cli --config <CONFIG_PATH> map put --name my-map --key b --value-type json --value '{"english":"Greetings"}'
hz-cli --config <CONFIG_PATH> map get --name my-map --key b
> {"english":"Greetings"}
```
