@echo off

setlocal
FOR /F "tokens=* USEBACKQ" %%F IN (`git rev-list --tags --max-count=1`) DO (
set GIT_COMMIT=%%F
)
FOR /F "tokens=* USEBACKQ" %%F IN (`git describe --tags %GIT_COMMIT%`) DO (
set CLC_VERSION=%%F
)
set ldflags="-X 'github.com/hazelcast/hazelcast-go-client/internal.ClientType=CLC' -X 'github.com/hazelcast/hazelcast-commandline-client/internal.GitCommit=%GIT_COMMIT%' -X 'github.com/hazelcast/hazelcast-commandline-client/internal.ClientVersion=%CLC_VERSION%' -X 'github.com/hazelcast/hazelcast-go-client/internal.ClientVersion=%CLC_VERSION%'"

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
