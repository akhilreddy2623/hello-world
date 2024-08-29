[CmdletBinding()]
Param(
    [Parameter(Mandatory = $true)]
    [String]$EnvironmentName,
    [Parameter(Mandatory = $false)]
    [String]$Rootfolder ="C:\Projects\pv-tm-collections\apigateway\APIGateway"
)

$BasicConfigPath="$Rootfolder\Apiproxyconfig\Apiproxyconfig-Basic.json"
$EnvironmentConfigPath="$Rootfolder\Apiproxyconfig\Apiproxyconfig-$EnvironmentName.json"

$wrpscriptRoot = Split-Path ($MyInvocation.MyCommand.Definition) -Parent
$wrpscriptRoot= "$wrpscriptRoot\APISIXDeploymentscripts\APIGW\Scripts"

Write-Host "$wrpscriptRoot"
Write-Host "$BasicConfigPath"
Write-Host "$EnvironmentConfigPath"

######################################################################################################
## Pre-Deployment Task
######################################################################################################

$pretask="$wrpscriptRoot\deployment-pretasks.ps1"
. $pretask `
    -BasicConfigPath $BasicConfigPath `
    -EnvironmentConfigPath $EnvironmentConfigPath `
    -Rootfolder $Rootfolder `
    -SYSTEMARTIFACTSDIRECTORY  $Rootfolder
######################################################################################################
## Deploy API Proxy
######################################################################################################
$deployAPIProxy="$wrpscriptRoot\deploy-APISIXproxy.ps1"
. $deployAPIProxy 