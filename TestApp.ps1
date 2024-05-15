# Create a log directory
$LOG_DIR = "output"
if (-not (Test-Path $LOG_DIR)) {
    New-Item -ItemType Directory -Path $LOG_DIR | Out-Null
}

# Removes all files in the log directory (if it is non empty)
if ((Get-ChildItem -Path $LOG_DIR).Count -gt 0) {
    Remove-Item -Path "$LOG_DIR\*" -Force -Recurse
}

# Kill orphaned processes
Get-NetTCPConnection -LocalPort 160 -ErrorAction SilentlyContinue | Where-Object {$_.OwningProcess -ne $null -and $_.OwningProcess -ne 0} | ForEach-Object { Stop-Process -Id $_.OwningProcess -Force }

# Runs Go program tests
& go test -v app_test.go

# Merges log files
##Get-ChildItem -Path $LOG_DIR -Recurse -Include "*.log" | Get-Content | Set-Content -Path "$LOG_DIR\completeLog.log" -Force

# Merges GoVector-specific log files
##Get-ChildItem -Path "$LOG_DIR\GoVector" -Recurse -Include "LogFile*" | Get-Content | Out-File -FilePath "$LOG_DIR\GoVector\temp_log.log" -Force
##$pattern = "(?<date>\d+) (?<host>\w*) (?<clock>.*)`r?`n(?<event>.*)`r?`n"
##$matchResult = Select-String -Path "$LOG_DIR\GoVector\temp_log.log" -Pattern $pattern -AllMatches
##$matchResult.Matches.Value | Out-File -FilePath "$LOG_DIR\GoVector\completeGoVectorLog.log" -Force
##Remove-Item -Path "$LOG_DIR\GoVector\temp_log.log" -Force
