@echo off
REM Attempt to stop watcher-node.exe processes
taskkill /IM watcher-node.exe /F || echo No watcher-node.exe processes found
