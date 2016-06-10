# Docker Version Manager PowerShell Wrapper
# Implemented as a POSIX-compliant function
# To use, source this script, `. dvm.ps1`, then type dvm help

function dvm() {
  $dvmDir = $env:DVM_DIR
  if( $dvmDir -eq $null ) {
    $dvmDir = "$env:USERPROFILE\.dvm"
  }

  if( !(Test-Path (Join-Path $dvmDir dvm-helper\dvm-helper.exe)) ) {
    $host.ui.WriteErrorLine("Installation corrupt: dvm-helper.exe is missing. Please reinstall dvm.")
    return 1
  }

  $dvmOutput = Join-Path $dvmDir .tmp\dvm-output.ps1
  if( Test-Path $dvmOutput ) {
    rm $dvmOutput
  }

  $env:DVM_DIR = $dvmDir

  $rawArgs = $MyInvocation.Line.Substring(3).Trim()
  $dvmCall = "$dvmDir\dvm-helper\dvm-helper.exe --shell powershell $rawArgs"
  iex $dvmCall

  if( Test-Path $dvmOutput ) {
    . $dvmOutput
  }
}
