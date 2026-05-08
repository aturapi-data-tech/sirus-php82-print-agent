# =============================================================================
# Sirus Print Agent — One-Click Installer
# =============================================================================
# Yang dilakukan script ini:
#   1. Cek service existing → stop & remove kalau ada
#   2. Copy agent .exe + config.json + bundled tools (NSSM, SumatraPDF) ke
#      C:\Program Files\SirusPrintAgent\
#   3. Register Windows Service via NSSM (auto-start at boot, auto-restart on crash)
#   4. Open firewall rule untuk port 9999 (opsional, loopback biasanya tidak perlu)
#   5. Start service + buka browser ke dashboard agent
# =============================================================================

param(
    [string]$InstallDir = "C:\Program Files\SirusPrintAgent"
)

# Verify admin
if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "ERROR: Run setup.bat sebagai Administrator (klik kanan -> Run as administrator)" -ForegroundColor Red
    Pause
    exit 1
}

$ScriptDir   = Split-Path -Parent $MyInvocation.MyCommand.Path
$svcName     = "SirusPrintAgent"
$NssmPath    = "$InstallDir\nssm.exe"
$SumatraPath = "$InstallDir\SumatraPDF.exe"
$AgentPath   = "$InstallDir\sirus-print-agent.exe"
$ConfigPath  = "$InstallDir\config.json"

Write-Host ""
Write-Host "===========================================" -ForegroundColor Cyan
Write-Host " Sirus Print Agent - One-Click Setup" -ForegroundColor Cyan
Write-Host "===========================================" -ForegroundColor Cyan
Write-Host ""

# --- 1. Stop & remove service existing ---
Write-Host "[1/5] Cek service existing..." -ForegroundColor Yellow
$existing = Get-Service $svcName -ErrorAction SilentlyContinue
if ($existing) {
    Write-Host "      Service ditemukan, stop & remove dulu..."
    if ($existing.Status -eq 'Running') {
        Stop-Service $svcName -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 1
    }
    if (Test-Path "$InstallDir\nssm.exe") {
        & "$InstallDir\nssm.exe" remove $svcName confirm | Out-Null
    } else {
        sc.exe delete $svcName | Out-Null
    }
    Start-Sleep -Seconds 1
    Write-Host "      OK" -ForegroundColor Green
} else {
    Write-Host "      Belum ada (fresh install)" -ForegroundColor Green
}

# --- 2. Setup folder ---
Write-Host "[2/5] Buat folder $InstallDir..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Write-Host "      OK" -ForegroundColor Green

# --- 3. Copy file ---
Write-Host "[3/5] Copy agent + tools (NSSM, SumatraPDF)..." -ForegroundColor Yellow
Copy-Item "$ScriptDir\sirus-print-agent.exe" $AgentPath -Force
Copy-Item "$ScriptDir\tools\nssm.exe" $NssmPath -Force
Copy-Item "$ScriptDir\tools\SumatraPDF.exe" $SumatraPath -Force

# Config: don't overwrite if exists (preserve customization)
if (Test-Path $ConfigPath) {
    Write-Host "      config.json sudah ada di $InstallDir - TIDAK overwrite (preserved)" -ForegroundColor Yellow
} else {
    Copy-Item "$ScriptDir\config.json" $ConfigPath -Force
    # Update path SumatraPDF di config jadi ke folder install (agent self-contained)
    $sumatraEscaped = $SumatraPath.Replace('\','\\')
    (Get-Content $ConfigPath) -replace 'C:\\\\Tools\\\\SumatraPDF\.exe', $sumatraEscaped | Set-Content $ConfigPath
    Write-Host "      config.json baru dibuat dengan SumatraPDF di $SumatraPath" -ForegroundColor Green
}

# --- 4. Install Windows Service via NSSM ---
Write-Host "[4/5] Install Windows Service..." -ForegroundColor Yellow
& $NssmPath install $svcName $AgentPath | Out-Null
& $NssmPath set $svcName AppDirectory $InstallDir | Out-Null
& $NssmPath set $svcName DisplayName "Sirus Print Agent" | Out-Null
& $NssmPath set $svcName Description "Local print agent untuk Sirus PHP web app - listen http://localhost:9999" | Out-Null
& $NssmPath set $svcName Start SERVICE_AUTO_START | Out-Null
& $NssmPath set $svcName AppStdout "$InstallDir\stdout.log" | Out-Null
& $NssmPath set $svcName AppStderr "$InstallDir\stderr.log" | Out-Null
& $NssmPath set $svcName AppStdoutCreationDisposition 4 | Out-Null
& $NssmPath set $svcName AppStderrCreationDisposition 4 | Out-Null
& $NssmPath set $svcName AppRotateFiles 1 | Out-Null
& $NssmPath set $svcName AppRotateOnline 1 | Out-Null
& $NssmPath set $svcName AppRotateBytes 10485760 | Out-Null
& $NssmPath set $svcName AppExit Default Restart | Out-Null
& $NssmPath set $svcName AppRestartDelay 2000 | Out-Null

# Firewall rule (loopback usually ga perlu, tapi safe-add)
netsh advfirewall firewall delete rule name="Sirus Print Agent" 2>$null | Out-Null
netsh advfirewall firewall add rule name="Sirus Print Agent" dir=in action=allow protocol=TCP localport=9999 profile=any | Out-Null
Write-Host "      OK" -ForegroundColor Green

# --- 5. Start service ---
Write-Host "[5/5] Starting service..." -ForegroundColor Yellow
Start-Service $svcName -ErrorAction SilentlyContinue
Start-Sleep -Seconds 2

# --- Verify ---
$svc = Get-Service $svcName -ErrorAction SilentlyContinue
Write-Host ""
if ($svc -and $svc.Status -eq 'Running') {
    Write-Host "===========================================" -ForegroundColor Green
    Write-Host " INSTALL SUKSES" -ForegroundColor Green
    Write-Host "===========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "  Service     : $($svc.Name) [$($svc.Status)]"
    Write-Host "  Folder      : $InstallDir"
    Write-Host "  Dashboard   : http://localhost:9999/"
    Write-Host "  Health check: http://localhost:9999/health"
    Write-Host "  Log file    : $InstallDir\stdout.log"
    Write-Host ""
    Write-Host "Edit $ConfigPath untuk sesuaikan printer name." -ForegroundColor Yellow
    Write-Host "Setelah edit, restart service:" -ForegroundColor Yellow
    Write-Host "  net stop SirusPrintAgent ; net start SirusPrintAgent" -ForegroundColor Yellow
    Write-Host ""

    # Try open browser to dashboard
    Start-Process "http://localhost:9999/"
} else {
    Write-Host "===========================================" -ForegroundColor Red
    Write-Host " GAGAL!" -ForegroundColor Red
    Write-Host "===========================================" -ForegroundColor Red
    Write-Host ""
    Write-Host "  Service tidak bisa start." -ForegroundColor Red
    Write-Host "  Cek log: $InstallDir\stderr.log"
    if ($svc) {
        Write-Host "  Status : $($svc.Status)"
    }
}

Write-Host ""
Pause
