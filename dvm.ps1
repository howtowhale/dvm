# Docker Version Manager PowerShell Wrapper
# Implemented as a POSIX-compliant function
# To use, source this script, `. dvm.ps1`, then type dvm help

function downloadDvm()
{
  $dvmDir = $PSScriptRoot
  if( Test-Path (Join-Path $dvmDir dvm-helper.exe) ) {
    return
  }

  if( [System.Environment]::Is64BitOperatingSystem ) { $arch = "amd64" } else { $arch = "386"}
  $latestTag = (ConvertFrom-Json (wget https://api.github.com/repos/getcarina/dvm/tags).Content) | select -ExpandProperty name | select -first 1
  (New-Object net.webclient).DownloadFile("https://github.com/getcarina/dvm/releases/download/$latestTag/dvm-helper-windows-$arch.exe", "$dvmDir\dvm-helper.exe")
}

function dvm()
{
  downloadDvm

  $dvmDir = $PSScriptRoot
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
