# Hazelcast CLC

## Requirements

* Go 1.15 or better

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
go build github.com/hazelcast/hazelcast-commandline-client
```

## Running

Make sure a Hazelcast v5.2021.07.1 instance is running.

```
# Get help
hzc --help

# Put to a map
hzc map put --name my-map --key my-key --value my-value

# Get from a map
hzc map get --name my-map --key my-key
```

## Examples

```
hzc map put --name my-map --key a --value-type string --value "Greetings"
hzc map get --name my-map --key a
> "Greetings"
hzc map put --name my-map --key b --value-type json --value {"english":"Greetings"}
hzc map get --name my-map --key b
> {english:Greetings}
```