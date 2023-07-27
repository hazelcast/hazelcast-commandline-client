# Simple Streaming Pipeline

Requirements:

1. JRE 8
2. Gradle 8

## Usage

### Compile the project

```
gradle shadowJar
```

### Submit the Jet Job

```
clc job submit ./build/libs/simple-streaming-pipeline-1.0-SNAPSHOT-all.jar
```

### Observe the Map Updates

```
clc map -n my-map entry-set
```

### Clean Up

```
clc job cancel ./build/libs/simple-streaming-pipeline-1.0-SNAPSHOT-all.jar
```