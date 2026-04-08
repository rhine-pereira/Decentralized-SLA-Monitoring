@echo off
REM Quick manual test script
REM 1) Run single node in background (use start or separate terminal)
REM 2) Wait 15 seconds
REM 3) Verify logs

echo Starting single node for test...
start "Test Watcher" cmd /k "go run . -config config_A.json"

echo Waiting 20 seconds to collect samples...
timeout /t 20 /nobreak >nul

echo Verifying logs...
go run . -run verify_cli -log logs\watcher_A.jsonl || go run verify_cli.go -log logs\watcher_A.jsonl

echo Test complete. Please inspect outputs.
