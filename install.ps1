$ErrorActionPreference = "Stop"

function installDvm()
{
  $dvmDir = $env:DVM_DIR
  if( $dvmDir -eq $null ) {
    $dvmDir = "$env:USERPROFILE\.dvm"
  }

  if( !(Test-Path $dvmDir) ) {
    mkdir $dvmDir > $null
  }

  $webClient = New-Object net.webclient

  echo "Downloading dvm.ps1..."
  $webClient.DownloadFile("https://rawgit.com/getcarina/dvm/master/dvm.ps1", "$dvmDir\dvm.ps1")

  echo "Downloading dvm.cmd..."
  $webClient.DownloadFile("https://cdn.rawgit.com/getcarina/dvm/master/dvm.cmd", "$dvmDir\dvm.cmd")

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
