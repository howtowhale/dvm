:: Docker Version Manager CMD Wrapper
:: Implemented as a POSIX-compliant function
:: To use, add this script's parent directory to your path

@ECHO OFF
SETLOCAL

SET DVM_DIR=%~dp0

IF NOT EXIST "%DVM_DIR%\dvm-helper\dvm-helper.exe" (
  echo Installation corrupt: dvm-helper.exe is missing. Please reinstall dvm.
  EXIT /b 1
)

SET DVM_OUTPUT="%DVM_DIR%\.tmp\dvm-output.cmd"

IF EXIST "%DVM_OUTPUT%" (
  DEL "%DVM_OUTPUT%"
)

%DVM_DIR%\dvm-helper\dvm-helper.exe --shell cmd %*

IF EXIST "%DVM_OUTPUT%" (
  ENDLOCAL
  CALL "%DVM_OUTPUT%"
)
