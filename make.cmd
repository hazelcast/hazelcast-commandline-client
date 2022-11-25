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

call :%1
if errorlevel 1 (
    echo Unknown target: %1
)

goto :end

:build
    go-winres make --product-version=%CLC_VERSION% --file-version=%CLC_VERSION%
    go build -tags base,hazelcastinternal,hazelcastinternaltest -ldflags %ldflags% -o clc.exe .
    goto :end

:installer
    call make.cmd build
    ISCC.exe /O%cd% /Fhazelcast-clc-setup-%CLC_VERSION% /DSourceDir=%cd% %cd%\extras\windows\installer\hazelcast-clc-installer.iss
    goto :end

:end
endlocal
