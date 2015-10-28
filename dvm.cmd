:: Docker Version Manager CMD Wrapper
:: Implemented as a POSIX-compliant function
:: To use, add this script's parent directory to your path

@ECHO OFF
SETLOCAL

SET DVM_DIR=%~dp0

IF NOT EXIST dvm-helper.exe (
  CALL powershell ". dvm.ps1; downloadDvm"
)

SET DVM_OUTPUT=%DVM_DIR%\.tmp\dvm-output.cmd

IF EXIST %DVM_OUTPUT% (
  DEL %DVM_OUTPUT%
)

%DVM_DIR%\dvm-helper.exe --shell cmd %*

IF EXIST %DVM_OUTPUT% (
  ENDLOCAL
  CALL %DVM_OUTPUT%
)
