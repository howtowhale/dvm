$ErrorActionPreference = "Stop"

function downloadDvm([string] $dvmDir) {
  $webClient = New-Object net.webclient

  echo "Downloading dvm.ps1..."
  $webClient.DownloadFile("https://raw.githubusercontent.com/getcarina/dvm/master/dvm.ps1", "$dvmDir\dvm.ps1")

  echo "Downloading dvm.cmd..."
  $webClient.DownloadFile("https://raw.githubusercontent.com/getcarina/dvm/master/dvm.cmd", "$dvmDir\dvm.cmd")

  echo "Downloading dvm-helper.exe..."
  $tmpDir = Join-Path $dvmDir .tmp
  if( !(Test-Path $tmpDir) ) {
    mkdir $tmpDir > $null
  }

  # Detect x86 vs. x64
  if( [System.Environment]::Is64BitOperatingSystem ) { $arch = "amd64" } else { $arch = "386"}

  # Download latest release
  $latestTag = (ConvertFrom-Json (wget https://api.github.com/repos/getcarina/dvm/tags).Content) | select -ExpandProperty name | select -first 1

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

function installDvm()
{
  $dvmDir = $env:DVM_DIR
  if( $dvmDir -eq $null ) {
    $dvmDir = "$env:USERPROFILE\.dvm"
  }

  if( !(Test-Path $dvmDir) ) {
    mkdir $dvmDir > $null
  }

  downloadDvm $dvmDir

  echo "Docker Version Manager (dvm) has been installed to $dvmDir"
  echo ""
  echo "PowerShell Users: Add the following command to your PowerShell profile:"
  echo "`t. $dvmDir\dvm.ps1"
  echo ""
  echo "CMD Users: Run the following commands to add dvm.cmd to your PATH:"
  echo "`tPATH=%PATH%;$dvmDir"
  echo "`tsetx PATH `"%PATH%;$dvmDir`""
}

installDvm
