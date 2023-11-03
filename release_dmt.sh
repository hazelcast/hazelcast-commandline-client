GOOS=linux GOARCH=amd64 BINARY_NAME=dmt_linux_amd64 MAIN_CMD_HELP='Hazelcast Data Migration Tool' make build-dmt
GOOS=linux GOARCH=arm64 BINARY_NAME=dmt_linux_arm64 MAIN_CMD_HELP='Hazelcast Data Migration Tool' make build-dmt
GOOS=darwin GOARCH=amd64 BINARY_NAME=dmt_darwin_amd64 MAIN_CMD_HELP='Hazelcast Data Migration Tool' make build-dmt
GOOS=darwin GOARCH=arm64 BINARY_NAME=dmt_darwin_arm64 MAIN_CMD_HELP='Hazelcast Data Migration Tool' make build-dmt
GOOS=windows GOARCH=amd64 BINARY_NAME=dmt_windows_amd64.exe MAIN_CMD_HELP='Hazelcast Data Migration Tool' make build-dmt

