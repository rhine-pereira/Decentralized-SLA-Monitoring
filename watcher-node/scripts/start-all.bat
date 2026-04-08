@echo off
REM Build binary
go build -o watcher-node.exe .

REM Start 5 nodes in separate windows (keep windows open)
start "Watcher A" cmd /k "watcher-node.exe -config config_A.json"
start "Watcher B" cmd /k "watcher-node.exe -config config_B.json"
start "Watcher C" cmd /k "watcher-node.exe -config config_C.json"
start "Watcher D" cmd /k "watcher-node.exe -config config_D.json"
start "Watcher E" cmd /k "watcher-node.exe -config config_E.json"

echo Launched 5 watcher nodes.
