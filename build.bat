set GOPATH=%CD%
go build -ldflags -Hwindowsgui teratogen
IF %ERRORLEVEL% NEQ 0 GOTO End
del assets.zip
tools\win\zip assets.zip assets\
type assets.zip>>teratogen.exe
tools\win\zip -A teratogen.exe
:End
