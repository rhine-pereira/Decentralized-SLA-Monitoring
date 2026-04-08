@echo off
REM Restructure directories script

cd /d "%~dp0"

echo Creating directory structure...
mkdir cmd\verify 2>nul
mkdir configs 2>nul
mkdir scripts 2>nul
mkdir docs 2>nul
mkdir logs 2>nul
mkdir keys 2>nul

echo Moving files to organized structure...

REM Move verify CLI
if exist verify_cli.go (
    move /Y verify_cli.go cmd\verify\main.go >nul
)

REM Move config files
if exist config_A.json move /Y config_A.json configs\ >nul
if exist config_B.json move /Y config_B.json configs\ >nul
if exist config_C.json move /Y config_C.json configs\ >nul
if exist config_D.json move /Y config_D.json configs\ >nul
if exist config_E.json move /Y config_E.json configs\ >nul
if exist example-config.json move /Y example-config.json configs\ >nul

REM Move scripts
if exist run-node.bat move /Y run-node.bat scripts\ >nul
if exist run-tests.bat move /Y run-tests.bat scripts\ >nul
if exist start-all.bat move /Y start-all.bat scripts\ >nul
if exist stop-all.bat move /Y stop-all.bat scripts\ >nul

REM Move docs
if exist sample_output.md move /Y sample_output.md docs\ >nul
if exist sample_watcher_A.jsonl move /Y sample_watcher_A.jsonl docs\ >nul

echo ✓ Directory structure organized
echo.
echo Structure:
echo   watcher-node/
echo   ├── cmd/verify/        (verification CLI)
echo   ├── configs/           (node configurations)
echo   ├── scripts/           (helper scripts)
echo   ├── docs/              (documentation)
echo   ├── logs/              (runtime logs)
echo   ├── keys/              (keypairs)
echo   └── [Go source files]  (main.go, types.go, etc.)
echo.
pause
