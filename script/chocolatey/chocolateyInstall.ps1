# Use the appropriate helper binary
$toolsDir = Split-Path $script:MyInvocation.MyCommand.Path
$helper   = Join-Path $toolsDir "dvm-helper/dvm-helper.exe"
$helper32 = Join-Path $toolsDir "dvm-helper/dvm-helper.x86.exe"
$helper64 = Join-Path $toolsDir "dvm-helper/dvm-helper.x64.exe"
if( $env:chocolateyForceX86 -or Get-ProcessorBits 32 ) { Copy-Item -Force $helper32 $helper }
else { Move-Item -Force $helper64 $helper }

# Install dvm function in the user's profile
$dvmScript = Join-Path $(Split-Path -Parent $MyInvocation.MyCommand.Definition) "dvm.ps1"
$dotSource = @"
# Load the Docker Version Manager (dvm)
. $dvmScript
"@
Add-Content $profile $dotSource

# Ensure dvm home directory exists
$dvmDir = if( $env:DVM_DIR -eq $null ) { "$env:USERPROFILE\.dvm" } else { $env:DVM_DIR }
if( !(Test-Path $dvmDir) ) { mkdir $dvmDir > $null }
