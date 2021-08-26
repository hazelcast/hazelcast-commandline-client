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
```
hz-cli map put --name my-map --key a --value-type string --value "Meet"
hz-cli map get --name my-map --key a
> "Meet"
hz-cli map put --name my-map --key b --value-type json --value '{"english":"Greetings"}'
hz-cli map get --name my-map --key b
> {"english":"Greetings"}
```
### Custom Configuration via Command Line
#### Connect to Hazelcast Cloud
```
hz-cli --cloud-token <YOUR_HAZELCAST_CLOUD_TOKEN> --cluster-name <YOUR_CLUSTER_NAME> map put --name map --key a --value-type json --value '{"meet":"greet"}'
hz-cli --cloud-token <YOUR_HAZELCAST_CLOUD_TOKEN> --cluster-name <YOUR_CLUSTER_NAME> map get --name map --key a
> {"meet":"greet"}
```

#### Connect to Local Hazelcast instance
```
hz-cli --address 192.168.1.1:5701,192.168.1.1:5702 --cluster-name <YOUR_CLUSTER_NAME> map put --name my-map --key a --value-type string --value "Meet"
hz-cli --address 192.168.1.1:5701,192.168.1.1:5702 --cluster-name <YOUR_CLUSTER_NAME> map get --name my-map --key a
> "Meet"
```

#### Use custom configuration file
```
hz-cli --config <CONFIG_PATH> mapp put --name my-map --key a --value-type string --value "Meet"
hz-cli --config <CONFIG_PATH> mapp get --name my-map --key a
> "Meet"
```