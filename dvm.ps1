# Docker Version Manager PowerShell Wrapper
# Implemented as a POSIX-compliant function
# To use, source this script, `. dvm.ps1`, then type dvm help

function dvm()
{
  $dvmDir = $PSScriptRoot
  if( !(Test-Path (Join-Path $dvmDir dvm-helper.exe)) )
  {
    # TODO: Download the binary instead of grabbing it from a local build
    cp $dvmDir\bin\windows\amd64\dvm-helper.exe $dvmDir
  }

  $dvmOutput = Join-Path $dvmDir .tmp\dvm-output.ps1
  if( Test-Path $dvmOutput )
  {
    rm $dvmOutput
  }

  $env:DVM_DIR = $dvmDir

  $rawArgs = $MyInvocation.Line.Substring(3).Trim()
  $dvmCall = "$dvmDir\dvm-helper.exe --shell powershell $rawArgs"
  iex $dvmCall

  if( Test-Path $dvmOutput )
  {
    . $dvmOutput
  }
}
