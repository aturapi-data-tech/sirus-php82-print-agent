# Sirus Print Agent

Local agent (Go) yang menjembatani web Sirus dengan printer fisik di PC user — auto-print etiket apotek & struk kasir tanpa popup.

Menggantikan pattern Oracle Forms `HOST(FoxitReader.exe /t pdf "printer", NO_SCREEN)` ke arsitektur web modern: agent listen di `http://localhost:9999`, browser fetch via JS.

## Konsep

```
[ Browser Sirus ] ──fetch──▶ [ http://localhost:9999/cetak-pdf ]
                                      │
                                      ▼ (sirus-print-agent.exe)
                              SumatraPDF.exe -print-to "etiket" -silent file.pdf
                                      │
                                      ▼
                              Printer "etiket" / "kasir" di PC user
```

**2 printer key default:**
- `etiket` — label apotek, gelang pasien, sticker info (thermal/dot-matrix)
- `kasir`  — struk/kwitansi pembayaran (thermal receipt)

## Stack

- **Go 1.18+** — single-binary, no runtime di PC user
- **SumatraPDF** (bundled) — silent PDF print, MIT license
- **NSSM** (bundled) — install agent jadi Windows Service

Bundle siap distribusi: **23 MB** (agent + SumatraPDF + NSSM + installer).

## Endpoints

| Method | Path | Fungsi |
|---|---|---|
| GET  | `/`                 | Dashboard HTML (debug + manual test) |
| GET  | `/health`           | Liveness check |
| GET  | `/printers`         | List logical printer key di config |
| GET  | `/system-printers`  | List printer Windows asli (via wmic) |
| POST | `/cetak-pdf`        | Cetak PDF — body: `{"pathPdf": "...", "printerKey": "etiket"}` |

## Build (di Linux/Mac dev)

```bash
# Pre-req: Go 1.18+, Tool/nssm.exe + Tool/SumatraPDF-3.6.1-32.exe ada di project
./build-windows.sh
# Output: dist/ (23 MB) — siap di-distribusi ke PC Windows
```

## Deploy ke PC Windows User

### Pre-req per PC (one-time setup):

**Rename printer di Windows ke nama seragam** (agar config standar bisa dipakai di semua PC):

```powershell
# Run as Administrator
Rename-Printer -Name "Argox OS-2140"   -NewName "etiket"
Rename-Printer -Name "Epson TM-T82III" -NewName "kasir"
```

(Ganti nama printer asli sesuai yang ada di PC tsb. Cek dengan `Get-Printer`)

### Install agent

1. Copy seluruh isi `dist/` ke PC user (USB/share/network)
2. **Klik kanan `setup.bat` → Run as administrator**
3. Tunggu "INSTALL SUKSES" → browser otomatis terbuka ke dashboard

Selesai. Service auto-start saat boot, auto-restart saat crash.

### Uninstall

Klik kanan `uninstall.bat` → Run as administrator.

## Cara Pakai dari Laravel

### Sample integrasi (Blade + JS)

```html
<x-primary-button x-on:click="cetakEtiket('{{ $etiketPdfUrl }}')">
    Cetak Etiket
</x-primary-button>

<x-primary-button x-on:click="cetakKasir('{{ $strukPdfUrl }}')">
    Cetak Struk
</x-primary-button>

<script>
async function cetakEtiket(pdfPath) {
    return cetakPDF(pdfPath, 'etiket');
}
async function cetakKasir(pdfPath) {
    return cetakPDF(pdfPath, 'kasir');
}

async function cetakPDF(pdfPath, printerKey) {
    try {
        const res = await fetch('http://localhost:9999/cetak-pdf', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ pathPdf: pdfPath, printerKey: printerKey })
        });
        const data = await res.json();
        if (data.ok) {
            Livewire.dispatch('toast', { type: 'success', message: data.msg });
        } else {
            Livewire.dispatch('toast', { type: 'error', message: data.msg });
        }
    } catch (e) {
        Livewire.dispatch('toast', { type: 'error',
            message: 'Print agent tidak aktif. Hubungi IT.' });
    }
}
</script>
```

### Path PDF source

`pathPdf` bisa berupa:
- **UNC share**: `\\\\<server>\\<share>\\file.pdf` (note: escape backslash di JSON)
- **Local PC user**: `C:\\path\\to\\file.pdf`
- **Mount path agent** (kalau PDF di-host Laravel): generate file ke share atau local temp dulu

## Konfigurasi

`C:\Program Files\SirusPrintAgent\config.json`:

```json
{
    "port": 9999,
    "pdfReaderPath": "C:\\Program Files\\SirusPrintAgent\\SumatraPDF.exe",
    "pdfReaderType": "sumatra",
    "allowedOrigins": [
        "http://localhost:8000",
        "http://<ip-server-laravel>",
        "http://sirus.local"
    ],
    "printers": {
        "etiket": "etiket",
        "kasir":  "kasir"
    }
}
```

**Penting:** `allowedOrigins` harus berisi origin web Sirus (CORS protection). Tambahkan domain/IP server Laravel di RS.

Setelah edit, restart service:
```cmd
net stop SirusPrintAgent
net start SirusPrintAgent
```

## File Structure

```
sirus-print-agent/
├── main.go, config.go, handlers.go, printer.go    Go source code
├── go.mod, config.json                            Module + default config
├── build-windows.sh                               Cross-compile + bundle script
├── installer/                                     Source installer scripts
│   ├── setup.bat, setup.ps1                      One-click installer
│   ├── uninstall.bat, uninstall.ps1
│   └── README-INSTALL.txt                         User-facing guide
├── Tool/                                          Bundled binaries source
│   ├── nssm.exe                                   (kept verbatim)
│   └── SumatraPDF-3.6.1-32.exe
├── dist/                                          Build output (siap distribusi)
│   ├── sirus-print-agent.exe                     Main agent
│   ├── config.json
│   ├── setup.bat, setup.ps1                      One-click install
│   ├── uninstall.bat, uninstall.ps1
│   ├── README-INSTALL.txt
│   └── tools/{nssm.exe, SumatraPDF.exe}
└── README.md                                      This file
```

## Development (Linux/Mac)

```bash
# Run dev server (port 9999)
go run .

# Test
curl http://localhost:9999/health
curl http://localhost:9999/printers
curl -X POST http://localhost:9999/cetak-pdf \
    -H "Content-Type: application/json" \
    -d '{"pathPdf":"/tmp/test.pdf","printerKey":"etiket"}'
```

Cross-compile ke Windows:
```bash
./build-windows.sh
# Output: dist/sirus-print-agent.exe (~4.5 MB)
```

Note: SumatraPDF tidak jalan di Linux — print command akan error. Test print full hanya di Windows.

## Arsitektur Detail

```
┌─────────────────────────┐
│  Server Laravel (PHP)   │
│  - Generate PDF         │
│  - Save ke share folder │ ───── HTTP fetch ─────▶  PC user (Windows)
│    \\<server>\<share>\  │                          ┌─────────────────────────┐
└─────────────────────────┘                          │ sirus-print-agent.exe   │
                                                      │ Windows Service         │
                                                      │ listen 127.0.0.1:9999   │
                                                      │   ↓ exec                │
                                                      │ SumatraPDF.exe          │
                                                      │ -print-to "etiket"      │
                                                      │ -silent "..."           │
                                                      │   ↓                     │
                                                      │ Printer USB lokal       │
                                                      │ (Argox/Epson dst)       │
                                                      └─────────────────────────┘
```

## Troubleshooting

Lihat `installer/README-INSTALL.txt` bagian Troubleshooting.

Quick check di PC user:
```cmd
sc query SirusPrintAgent              :: cek service status
type "C:\Program Files\SirusPrintAgent\stderr.log"   :: cek error log
"C:\Program Files\SirusPrintAgent\SumatraPDF.exe" -print-to "etiket" -silent "C:\test.pdf"   :: test manual
```

## Lisensi

Internal RS Sirus. Tools yang bundled:
- SumatraPDF — GPL v3
- NSSM — Public Domain
