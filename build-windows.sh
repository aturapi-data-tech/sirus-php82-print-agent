#!/bin/bash
# Build agent + bundle one-click installer untuk Windows.
#
# Output: dist/  — copy folder ini ke PC Windows, klik kanan setup.bat → Run as admin.
#
# Pre-req:
#   - Go 1.18+ ter-install (cek: go version)
#   - Tool/nssm.exe + Tool/SumatraPDF-3.6.1-32.exe sudah ada di project
#
# Pemakaian: ./build-windows.sh

set -e

DIST=dist
TOOL_NSSM=Tool/nssm.exe
TOOL_SUMATRA=Tool/SumatraPDF-3.6.1-32.exe

# Validasi tools
if [[ ! -f "$TOOL_NSSM" ]]; then
    echo "ERROR: $TOOL_NSSM tidak ditemukan."
    exit 1
fi
if [[ ! -f "$TOOL_SUMATRA" ]]; then
    echo "ERROR: $TOOL_SUMATRA tidak ditemukan."
    exit 1
fi

# Cleanup dist sebelumnya, biar tidak ada file lama nyangkut
rm -rf "$DIST"
mkdir -p "$DIST/tools"

echo "[1/4] Build Go agent for Windows AMD64..."
GOOS=windows GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o "$DIST/sirus-print-agent.exe" \
    .

echo "[2/4] Bundle tools (NSSM + SumatraPDF)..."
cp "$TOOL_NSSM" "$DIST/tools/nssm.exe"
cp "$TOOL_SUMATRA" "$DIST/tools/SumatraPDF.exe"

echo "[3/4] Bundle config + installer scripts..."
cp config.json "$DIST/config.json"
cp installer/setup.bat "$DIST/setup.bat"
cp installer/setup.ps1 "$DIST/setup.ps1"
cp installer/uninstall.bat "$DIST/uninstall.bat"
cp installer/uninstall.ps1 "$DIST/uninstall.ps1"
cp installer/README-INSTALL.txt "$DIST/README-INSTALL.txt"

echo "[4/4] Done!"
echo ""
echo "===== Bundle siap distribusi: $DIST/ ====="
ls -lh "$DIST/" "$DIST/tools/"
echo ""
echo "TOTAL UKURAN:"
du -sh "$DIST/"
echo ""
echo "Cara deploy ke PC Windows:"
echo "  1. Copy seluruh isi $DIST/ ke PC user (USB, network share, dll)"
echo "  2. Klik kanan setup.bat → Run as administrator"
echo "  3. Tunggu 'INSTALL SUKSES' → otomatis buka browser"
echo "  4. Edit C:\\Program Files\\SirusPrintAgent\\config.json sesuaikan printer"
