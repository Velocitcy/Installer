$link = "https://github.com/Velocitcy/Installer/releases/latest/download/VelocityInstallerCli.exe"

$outfile = "$env:TEMP\VelocityInstallerCli.exe"

Write-Output "Downloading installer to $outfile"

Invoke-WebRequest -Uri "$link" -OutFile "$outfile"

Write-Output ""

Start-Process -Wait -NoNewWindow -FilePath "$outfile"

# Cleanup
Remove-Item -Force "$outfile"
