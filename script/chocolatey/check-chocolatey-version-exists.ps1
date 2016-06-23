param (
  [string]$version
)

try {
  $url = "http://chocolatey.org/packages/carina/$version"
  $status = (Invoke-WebRequest $url).StatusCode
} catch {
  $status = $_.Exception.Response.StatusCode.Value__
}

$shouldPublishToNuGet = $status -eq '404'
Write-Host $shouldPublishToNuGet
