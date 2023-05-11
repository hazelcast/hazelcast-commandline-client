# Sample Jet Job with Dependencies

## Build

Requirements:

1. Java 11 or better
2. Gradle 8.0 or better (may work with previous versions)

```
$ ./gradlew shadowJar
```

That will create the `./build/libs/sample-job-1-1.0-SNAPSHOT-all.jar` file.
Copy that file to `base/commands/job/testdata` folder.