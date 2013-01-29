set GOPATH=%CD%
go run src/gen-version/gen-version.go
go build -ldflags -Hwindowsgui teratogen
IF %ERRORLEVEL% NEQ 0 GOTO End
del assets.zip
tools\win\zip assets.zip assets\
type assets.zip>>teratogen.exe
tools\win\zip -A teratogen.exe
:End
