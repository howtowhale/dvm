$ErrorActionPreference = "Stop"

function downloadDvm([string] $dvmDir) {
  $webClient = New-Object net.webclient

  echo "Downloading dvm.ps1..."
  $webClient.DownloadFile("https://download.getcarina.com/dvm/latest/dvm.ps1", "$dvmDir\dvm.ps1")

  echo "Downloading dvm.cmd..."
  $webClient.DownloadFile("https://download.getcarina.com/dvm/latest/dvm.cmd", "$dvmDir\dvm.cmd")

  echo "Downloading dvm-helper.exe..."
  $tmpDir = Join-Path $dvmDir .tmp
  if( !(Test-Path $tmpDir) ) {
    mkdir $tmpDir > $null
  }

  # Ensure dvm-helper directory exists
  $dvmHelperDir = Join-Path $dvmDir dvm-helper
  if( !(Test-Path $dvmHelperDir ) ) {
    mkdir $dvmHelperDir > $null
  }


  # Detect x86 vs. x64
  if( [System.Environment]::Is64BitOperatingSystem ) { $arch = "x86_64" } else { $arch = "i386"}

  # Download latest release
  $webClient.DownloadFile("https://download.getcarina.com/dvm/latest/Windows/$arch/dvm-helper.exe", "$tmpDir\dvm-helper.exe")
  $webClient.DownloadFile("https://download.getcarina.com/dvm/latest/Windows/$arch/dvm-helper.exe.sha256", "$tmpDir\dvm-helper.exe.256")

  # Verify the binary was downloaded successfully
  $checksum = (cat $tmpDir\dvm-helper.exe.256).Split(' ')[0]
  $hash = (Get-FileHash $tmpDir\dvm-helper.exe).Hash
  if([string]::Compare($checksum, $hash, $true) -ne 0) {
    $host.ui.WriteErrorLine("DANGER! The downloaded dvm-helper binary, $tmpDir\dvm-helper.exe, does not match its checksum!")
    $host.ui.WriteErrorLine("CHECKSUM: $checksum")
    $host.ui.WriteErrorLine("HASH: $hash")
    return 1
  }

  mv -force "$tmpDir\dvm-helper.exe" "$dvmDir\dvm-helper\dvm-helper.exe"
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

  echo ""
  echo "Docker Version Manager (dvm) has been installed to $dvmDir"
  echo ""
  echo "PowerShell Users: Run the following command to start using dvm. Then add it to your PowerShell profile to complete the installation."
  echo "`t. '$dvmDir\dvm.ps1'"
  echo ""
  echo "CMD Users: Run the first command to start using dvm. Then run the second command to add dvm to your PATH to complete the installation."
  echo "`t1. PATH=%PATH%;$dvmDir"
  echo "`t2. setx PATH `"%PATH%;$dvmDir`""
}

installDvm
