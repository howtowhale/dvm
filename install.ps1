function installMagic()
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

if($PSVersionTable -eq $null -or $PSVersionTable.PSVersion.Major -lt 4){
  Write-Output "magic requires PowerShell version 4 or higher."
  Exit 1
}

iwr -Uri https:// -OutFile $output
