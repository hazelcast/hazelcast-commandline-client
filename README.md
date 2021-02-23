# Hazelcast CLC

## Requirements

* Go 1.15 or better

## Build

```
git clone git@github.com:yuce/hzc.git
cd hzc
go build github.com/hazelcast/hzc/cmd/hzc
```

## Running

Make sure a Hazelcast v4 instance is running.  

```
# Put to a map
hzc map put --name my-map --key my-key --value my-value

# Get from a map
hzc map get --name my-map --key my-key
```