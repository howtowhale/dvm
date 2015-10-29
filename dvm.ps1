# Docker Version Manager PowerShell Wrapper
# Implemented as a POSIX-compliant function
# To use, source this script, `. dvm.ps1`, then type dvm help

function downloadDvm() {
  $dvmDir = $PSScriptRoot
  if( Test-Path (Join-Path $dvmDir dvm-helper.exe) ) {
    return
  }

  $tmpDir = Join-Path $dvmDir .tmp
  if( !(Test-Path $tmpDir) ) {
    mkdir $tmpDir > $null
  }

  # Detect mac vs. linux and x86 vs. x64
  if( [System.Environment]::Is64BitOperatingSystem ) { $arch = "amd64" } else { $arch = "386"}

  # Download latest release
  $latestTag = (ConvertFrom-Json (wget https://api.github.com/repos/getcarina/dvm/tags).Content) | select -ExpandProperty name | select -first 1
  $webClient = New-Object net.webclient
  $webClient.DownloadFile("https://github.com/getcarina/dvm/releases/download/$latestTag/dvm-helper-windows-$arch.exe", "$tmpDir\dvm-helper-windows-$arch.exe")
  $webClient.DownloadFile("https://github.com/getcarina/dvm/releases/download/$latestTag/dvm-helper-windows-$arch.exe.sha256", "$tmpDir\dvm-helper-windows-$arch.exe.256")

  # Verify the binary was downloaded successfully
  $checksum = (cat $tmpDir\dvm-helper-windows-$arch.exe.256).Split(' ')[0]
  $hash = (Get-FileHash $tmpDir\dvm-helper-windows-$arch.exe).Hash
  if([string]::Compare($checksum, $hash, $true) -ne 0) {
    $host.ui.WriteErrorLine("DANGER! The downloaded dvm-helper binary, $tmpDir\dvm-helper-windows-$arch.exe, does not match its checksum!")
    $host.ui.WriteErrorLine("CHECKSUM: $checksum")
    $host.ui.WriteErrorLine("HASH: $hash")
    return 1
  }

  mv "$tmpDir\dvm-helper-windows-$arch.exe" "$dvmDir\dvm-helper.exe"
}

function dvm() {
  if(downloadDvm -ne 0) {
    return 1
  }

  $dvmDir = $PSScriptRoot
  $dvmOutput = Join-Path $dvmDir .tmp\dvm-output.ps1
  if( Test-Path $dvmOutput ) {
    rm $dvmOutput
  }

  $env:DVM_DIR = $dvmDir

  $rawArgs = $MyInvocation.Line.Substring(3).Trim()
  $dvmCall = "$dvmDir\dvm-helper.exe --shell powershell $rawArgs"
  iex $dvmCall

  if( Test-Path $dvmOutput ) {
    . $dvmOutput
  }
}
