@echo off
setlocal

set OUT=out
set EXE=%OUT%\CronShot.exe
if not exist %OUT% mkdir %OUT%

echo [1/3] Generate Windows icon resources (optional)
for %%I in (rsrc.exe) do (
  if "%%~$PATH:I"=="" (
    echo - rsrc.exe not found, trying: go install github.com/akavel/rsrc@latest
    go install github.com/akavel/rsrc@latest
  )
)
rsrc -ico assets\icons\cat.ico -o resource.syso 2>nul
if errorlevel 1 (
  echo - rsrc failed, building without resource.syso (file icon may not be set)
)

echo [2/3] Build release exe with size optimizations
go build -trimpath -ldflags "-s -w -H=windowsgui" -tags release -o %EXE% .\
if errorlevel 1 (
  echo Build failed
  exit /b 1
)

echo [3/3] Optional UPX compress if available
for %%I in (upx.exe) do (
  if not "%%~$PATH:I"=="" (
    upx --best --lzma %EXE%
  )
)

rem cleanup temporary resource file
if exist resource.syso del /f /q resource.syso

echo Done. Output: %EXE%
endlocal
