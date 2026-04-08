@echo off
REM Usage: run-node.bat config_A.json
set CONFIG=%1
if "%CONFIG%"=="" set CONFIG=config_A.json
REM Build binary (uncomment if you prefer running built binary)
REM go build -o watcher-node.exe .

go run . -config %CONFIG%
