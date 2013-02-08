set GOPATH=%CD%
go run src/gen-version/gen-version.go
go install -ldflags -Hwindowsgui teratogen
IF %ERRORLEVEL% NEQ 0 GOTO End
del assets.zip
tools\win\zip -r assets.zip assets\
copy /b bin\teratogen.exe+assets.zip teratogen.exe
tools\win\zip -A teratogen.exe
:End
