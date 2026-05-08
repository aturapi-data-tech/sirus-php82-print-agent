# =============================================================================
# Sirus Print Agent — Uninstaller
# =============================================================================

param(
    [string]$InstallDir = "C:\Program Files\SirusPrintAgent"
)

if (-NOT ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
    Write-Host "ERROR: Run uninstall.bat sebagai Administrator" -ForegroundColor Red
    Pause
    exit 1
}

$svcName = "SirusPrintAgent"
$NssmPath = "$InstallDir\nssm.exe"

Write-Host ""
Write-Host "===========================================" -ForegroundColor Yellow
Write-Host " Sirus Print Agent - Uninstaller" -ForegroundColor Yellow
Write-Host "===========================================" -ForegroundColor Yellow
Write-Host ""

# Stop & remove service
Write-Host "[1/3] Stop & remove service..." -ForegroundColor Yellow
$svc = Get-Service $svcName -ErrorAction SilentlyContinue
if ($svc) {
    if ($svc.Status -eq 'Running') {
        Stop-Service $svcName -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 1
    }
    if (Test-Path $NssmPath) {
        & $NssmPath remove $svcName confirm | Out-Null
    } else {
        sc.exe delete $svcName | Out-Null
    }
    Write-Host "      OK" -ForegroundColor Green
} else {
    Write-Host "      Service tidak ditemukan" -ForegroundColor Yellow
}

# Remove firewall rule
Write-Host "[2/3] Remove firewall rule..." -ForegroundColor Yellow
netsh advfirewall firewall delete rule name="Sirus Print Agent" 2>$null | Out-Null
Write-Host "      OK" -ForegroundColor Green

# Optional: remove folder
Write-Host "[3/3] Hapus folder $InstallDir ?" -ForegroundColor Yellow
$reply = Read-Host "      Ketik Y untuk hapus, lainnya untuk skip"
if ($reply -eq 'Y' -or $reply -eq 'y') {
    if (Test-Path $InstallDir) {
        Remove-Item -Path $InstallDir -Recurse -Force -ErrorAction SilentlyContinue
        Write-Host "      Folder dihapus" -ForegroundColor Green
    }
} else {
    Write-Host "      Skip (folder dipertahankan: $InstallDir)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Uninstall selesai." -ForegroundColor Green
Write-Host ""
Pause
