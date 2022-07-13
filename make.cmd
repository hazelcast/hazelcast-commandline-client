@echo off

setlocal
set client_type=CLC
set client_version=v5.2.0-beta2
set ldflags="-X 'github.com/hazelcast/hazelcast-go-client/internal.ClientType=%client_type%' -X 'github.com/hazelcast/hazelcast-go-client/internal.ClientVersion=%client_version%'"

REM default target is build
if "%1" == "" (
    goto :build
)

2>NUL call :%1
if errorlevel 1 (
    echo Unknown target: %1
)

goto :end

:build
    go build -ldflags %ldflags% -o hzc.exe .
    goto :end

:end