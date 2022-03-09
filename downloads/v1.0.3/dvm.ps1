# Docker Version Manager PowerShell Wrapper
# Implemented as a POSIX-compliant function
# To use, source this script, `. dvm.ps1`, then type dvm help

function dvm() {
  # Default dvm home to ~/.dvm if not set
  $dvmDir = $env:DVM_DIR
  if( $dvmDir -eq $null ) {
    $dvmDir = "$env:USERPROFILE\.dvm"
  }

  # Expect that dvm-helper is next to this script
  $dvmHelper = Join-Path $PSScriptRoot dvm-helper\dvm-helper.exe
  if( !(Test-Path $dvmHelper) ) {
    $host.ui.WriteErrorLine("Installation corrupt: dvm-helper.exe is missing. Please reinstall dvm.")
    return 1
  }

  # Pass dvm-helper output back to script via ~/.dvm/.tmp/dvm-output.ps1
  $dvmOutput = Join-Path $dvmDir .tmp\dvm-output.ps1
  if( Test-Path $dvmOutput ) {
    rm $dvmOutput
  }

  $rawArgs = $MyInvocation.Line.Substring(3).Trim()
  $dvmCall = "& '$dvmHelper' --shell powershell $rawArgs"
  iex $dvmCall

  if( Test-Path $dvmOutput ) {
    . $dvmOutput
  }
}
