============================================================
  SIRUS PRINT AGENT - One-Click Installer
  Auto-print etiket apotek & struk kasir dari web Sirus
============================================================

KONSEP:
-------
  - Printer "etiket" (label/sticker) dan "kasir" (struk thermal)
    adalah printer LOKAL di PC user.
  - Web Sirus tinggal kirim perintah: print ke "etiket" atau "kasir"
    -> agent jalanin SumatraPDF -> langsung cetak tanpa popup.
  - Default printer Windows TIDAK terganggu.


============================================================
CARA INSTALL DI PC USER (1 menit, sekali setup)
============================================================

LANGKAH 1: Rename printer di Windows
-------------------------------------
  Penting: agar config standar bisa dipakai di semua PC tanpa edit.

  Cara via Settings:
    1. Buka Settings -> Bluetooth & devices -> Printers & scanners
    2. Klik printer LABEL (Argox/TSC/dll) -> klik nama -> klik
       "Printer properties" -> ubah nama jadi: etiket
    3. Klik printer STRUK THERMAL (Epson/Bixolon/dll) -> ubah jadi: kasir

  Cara cepat via PowerShell (Run as Administrator):
    Rename-Printer -Name "Argox OS-2140"   -NewName "etiket"
    Rename-Printer -Name "Epson TM-T82III" -NewName "kasir"

  (Ganti "Argox OS-2140" / "Epson TM-T82III" sesuai nama asli printer
   di PC ini. Cek dengan: Get-Printer | Select Name)


LANGKAH 2: Install agent
-------------------------
  1. Copy SELURUH ISI folder ini ke PC Windows (USB/share/dll).
  2. Klik kanan setup.bat -> "Run as administrator"
     (UAC prompt -> klik Yes)
  3. Tunggu sampai muncul "INSTALL SUKSES"
     -> Browser otomatis terbuka ke http://localhost:9999/

  Done! Service "SirusPrintAgent" auto-start saat Windows boot.


LANGKAH 3: Verifikasi
----------------------
  1. Di dashboard http://localhost:9999/, klik tombol:
     - "Cek Health"             -> harus return "ok: true"
     - "Cek Printers (Windows)" -> harus muncul "etiket" dan "kasir"
                                    di list
     - "Cek Printers (Config)"  -> harus muncul mapping etiket+kasir

  2. Test cetak PDF:
     - Path PDF: C:\Users\Public\sample.pdf  (atau PDF apapun)
     - Printer Key: etiket   (atau kasir)
     - Klik "Cetak Sekarang"
     - Output JSON harus "ok: true" + printer dapat job


============================================================
CARA PAKAI DARI WEB SIRUS / LARAVEL
============================================================

JavaScript (di Blade Laravel):
------------------------------
    async function cetakEtiket(pdfPath) {
        const res = await fetch('http://localhost:9999/cetak-pdf', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                pathPdf: pdfPath,           // path PDF (UNC or local)
                printerKey: 'etiket'        // 'etiket' atau 'kasir'
            })
        });
        const data = await res.json();
        if (data.ok) console.log('Print sukses:', data.msg);
        else         console.error('Print gagal:', data.msg);
    }

PDF source bisa berupa:
  - UNC path:  \\<server>\<share>\file.pdf
  - Local:     C:\path\to\file.pdf
  - Atau download dulu dari Laravel ke local sebelum print


============================================================
KONFIGURASI (kalau perlu custom)
============================================================

Edit file: C:\Program Files\SirusPrintAgent\config.json

Default config:
{
  "port": 9999,
  "pdfReaderPath": "C:\\Program Files\\SirusPrintAgent\\SumatraPDF.exe",
  "pdfReaderType": "sumatra",
  "allowedOrigins": ["http://localhost:8000", "http://server-laravel.local"],
  "printers": {
    "etiket": "etiket",     <-- ganti kalau printer Windows tidak rename
    "kasir":  "kasir"
  }
}

Setelah edit config, restart service:
  net stop SirusPrintAgent
  net start SirusPrintAgent


============================================================
CARA UNINSTALL
============================================================

Klik kanan uninstall.bat -> Run as administrator
- Service dihentikan + dihapus
- Folder optional dihapus (akan ditanya Y/N)


============================================================
YANG TER-INSTALL
============================================================

Folder: C:\Program Files\SirusPrintAgent\
  - sirus-print-agent.exe   (4.5 MB) - main agent
  - nssm.exe                (300 KB) - Windows service manager
  - SumatraPDF.exe          (18 MB)  - silent PDF print
  - config.json                      - mapping printer
  - stdout.log, stderr.log           - log file (auto-rotate 10 MB)

Windows Service:
  - Nama: SirusPrintAgent
  - Display: Sirus Print Agent
  - Start: Automatic (auto-start saat boot)
  - On crash: Auto-restart setelah 2 detik

Firewall:
  - Inbound TCP port 9999 (loopback)


============================================================
URL YANG TERSEDIA
============================================================

http://localhost:9999/                  Dashboard agent (HTML)
http://localhost:9999/health            Liveness check (JSON)
http://localhost:9999/printers          List printer key di config
http://localhost:9999/system-printers   List printer Windows asli
http://localhost:9999/cetak-pdf         POST endpoint cetak PDF


============================================================
TROUBLESHOOTING
============================================================

Q: setup.bat tutup setelah klik Yes UAC, tidak muncul installer
A: PowerShell Execution Policy mungkin disabled. Run as admin:
   Set-ExecutionPolicy RemoteSigned -Scope LocalMachine

Q: Service tidak running setelah install
A: 1. Cek log: C:\Program Files\SirusPrintAgent\stderr.log
   2. Coba run agent manual: cmd ke folder install -> sirus-print-agent.exe
   3. Cek: services.msc -> SirusPrintAgent -> Status

Q: Print tidak keluar tapi response sukses
A: 1. Test SumatraPDF manual di CMD:
      "C:\Program Files\SirusPrintAgent\SumatraPDF.exe" -print-to "etiket" -silent "C:\test.pdf"
   2. Pastikan printer "etiket" sudah di-rename benar
   3. Cek printer queue Windows: ada job stuck?

Q: Browser fetch ke http://localhost:9999 dapat CORS error
A: Tambah origin Laravel ke "allowedOrigins" di config.json
   Restart service setelah edit

Q: Service jalan tapi browser bilang "not found"
A: Pastikan akses ke "/" (root) bukan "/index.html". Dashboard ada di
   http://localhost:9999/ - kalau still 404 mungkin agent versi lama.
   Re-install dengan setup.bat (idempotent, aman re-run).

Q: Multiple PC, mau deploy massal
A: 1. Bundle dist/ ke shared folder atau USB
   2. Di tiap PC: rename printer ke "etiket" + "kasir"
   3. Run setup.bat -> selesai
   4. Config sama untuk semua PC, tidak perlu edit per-PC


============================================================
KONTAK
============================================================
  IT Support RS
============================================================
