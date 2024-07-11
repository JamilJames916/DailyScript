# To get the names and AUMIDs for Microsoft Store apps installed for the current user.

$installedapps = Get-AppxPackage

$aumidList = @()
foreach ($app in $installedapps)
{
    foreach ($id in (Get-AppxPackageManifest $app).package.applications.application.id)
    {
        $aumidList += $app.packagefamilyname + "!" + $id
    }
}

$aumidList


# You can add the -user <username> or the 
# -allusers parameters to the Get-AppxPackage cmdlet to list AUMIDs for other users. 
# You must use an elevated Windows PowerShell prompt to use the -user or -allusers parameters.