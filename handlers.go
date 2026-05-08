package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type APIResponse struct {
	OK  bool   `json:"ok"`
	Msg string `json:"msg,omitempty"`
}

type CetakRequest struct {
	PathPDF    string `json:"pathPdf"`     // opsional: path file (UNC / lokal)
	PdfBase64  string `json:"pdfBase64"`   // opsional: PDF content base64
	Filename   string `json:"filename"`    // opsional: nama file (untuk temp & log)
	PrinterKey string `json:"printerKey"`
}

// GET / — landing page dengan dashboard + tombol test (untuk admin/IT debug)
func handleIndex(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Hanya layani path root persis '/'
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, indexHTML,
			agentName,            // %s : <h1> nama
			agentVersion,         // %s : badge versi
			config.Port,          // %d : URL listening
			config.Port,          // %d : kv Port
			config.PdfReaderPath, // %s : kv PDF path
			config.PdfReaderType, // %s : kv reader type
			len(config.Printers), // %d : kv jumlah printer
			config.Port,          // %d : sample fetch URL di JS section
		)
	}
}

const indexHTML = `<!doctype html>
<html lang="id"><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Sirus Print Agent</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
       background: #f5f7fa; color: #1f2937; padding: 2rem; line-height: 1.5; }
.container { max-width: 900px; margin: 0 auto; }
h1 { font-size: 1.75rem; margin-bottom: 0.25rem; color: #0f172a; }
.subtitle { color: #64748b; margin-bottom: 2rem; }
.card { background: #fff; border: 1px solid #e2e8f0; border-radius: 12px;
        padding: 1.25rem; margin-bottom: 1rem; box-shadow: 0 1px 3px rgba(0,0,0,0.05); }
h2 { font-size: 1.1rem; margin-bottom: 0.75rem; color: #334155; }
.kv { display: grid; grid-template-columns: 200px 1fr; gap: 0.5rem 1rem; font-size: 0.9rem; }
.kv .k { color: #64748b; }
.kv .v { font-family: ui-monospace, "SF Mono", Menlo, Consolas, monospace; }
.badge { display: inline-block; padding: 0.15rem 0.55rem; border-radius: 999px;
         font-size: 0.75rem; font-weight: 600; }
.badge.ok { background: #dcfce7; color: #15803d; }
.endpoint { display: grid; grid-template-columns: 60px 1fr; gap: 0.75rem;
            padding: 0.6rem 0; border-bottom: 1px solid #f1f5f9; align-items: center; }
.endpoint:last-child { border-bottom: 0; }
.method { font-size: 0.75rem; font-weight: 700; padding: 0.15rem 0.5rem;
          border-radius: 4px; text-align: center; font-family: ui-monospace, monospace; }
.m-get { background: #dbeafe; color: #1e40af; }
.m-post { background: #fef3c7; color: #92400e; }
.path { font-family: ui-monospace, monospace; font-size: 0.85rem; color: #0f172a; }
.desc { font-size: 0.8rem; color: #64748b; margin-top: 0.15rem; }
button { font-family: inherit; cursor: pointer; padding: 0.4rem 0.85rem;
         border: 1px solid #cbd5e1; background: #fff; border-radius: 6px;
         font-size: 0.8rem; color: #334155; }
button:hover { background: #f1f5f9; }
button.primary { background: #16a34a; border-color: #16a34a; color: #fff; }
button.primary:hover { background: #15803d; }
.actions { display: flex; gap: 0.5rem; flex-wrap: wrap; margin-top: 0.75rem; }
pre { background: #0f172a; color: #e2e8f0; padding: 1rem; border-radius: 8px;
      font-size: 0.78rem; overflow-x: auto; font-family: ui-monospace, monospace; }
.output { background: #0f172a; color: #e2e8f0; padding: 1rem; border-radius: 8px;
          font-size: 0.78rem; min-height: 80px; max-height: 300px; overflow-y: auto;
          font-family: ui-monospace, monospace; white-space: pre-wrap; word-break: break-all; }
input[type=text] { font-family: ui-monospace, monospace; font-size: 0.85rem;
                   padding: 0.4rem 0.6rem; border: 1px solid #cbd5e1; border-radius: 6px;
                   width: 100%%; }
label { display: block; font-size: 0.8rem; font-weight: 600; margin: 0.5rem 0 0.25rem;
        color: #334155; }
.row { display: grid; grid-template-columns: 1fr 1fr; gap: 0.75rem; }
.note { background: #fef3c7; border-left: 4px solid #f59e0b; padding: 0.75rem 1rem;
        border-radius: 6px; font-size: 0.85rem; color: #78350f; margin-bottom: 1rem; }
</style>
</head><body><div class="container">

<h1>%s <span class="badge ok">v%s — RUNNING</span></h1>
<p class="subtitle">Local print agent — listening at <code>http://127.0.0.1:%d</code></p>

<div class="note">
<strong>Halaman ini cuma dashboard debug.</strong> Agent dipanggil dari Laravel/Sirus via JavaScript fetch ke endpoint di bawah. Jangan akses dari URL bar di production.
</div>

<div class="card">
<h2>Konfigurasi Aktif</h2>
<div class="kv">
<div class="k">Port</div><div class="v">%d</div>
<div class="k">PDF Reader Path</div><div class="v">%s</div>
<div class="k">PDF Reader Type</div><div class="v">%s</div>
<div class="k">Printer terdaftar</div><div class="v">%d (klik tombol "Cek Printers" di bawah)</div>
</div>
</div>

<div class="card">
<h2>Test Endpoint (klik untuk panggil)</h2>
<div class="endpoint">
<span class="method m-get">GET</span>
<div>
<span class="path">/health</span>
<div class="desc">Liveness check — return JSON status agent</div>
</div>
</div>
<div class="endpoint">
<span class="method m-get">GET</span>
<div>
<span class="path">/printers</span>
<div class="desc">List logical printer key di config.json (etiket, rm, kasir, ...)</div>
</div>
</div>
<div class="endpoint">
<span class="method m-get">GET</span>
<div>
<span class="path">/system-printers</span>
<div class="desc">List printer Windows asli (via wmic) — untuk lihat nama yang Windows kenali</div>
</div>
</div>
<div class="endpoint">
<span class="method m-post">POST</span>
<div>
<span class="path">/cetak-pdf</span>
<div class="desc">Cetak PDF — body: <code>{"pathPdf":"...", "printerKey":"etiket"}</code></div>
</div>
</div>

<div class="actions">
<button onclick="callApi('GET', '/health')">Cek Health</button>
<button onclick="callApi('GET', '/printers')">Cek Printers (Config)</button>
<button onclick="callApi('GET', '/system-printers')">Cek Printers (Windows)</button>
</div>
</div>

<div class="card">
<h2>Test Cetak PDF</h2>
<div class="row">
<div>
<label>Path PDF (di PC ini, atau UNC \\\\server\\share\\file.pdf)</label>
<input type="text" id="pathPdf" placeholder="C:\path\to\file.pdf  atau  \\server\share\file.pdf">
</div>
<div>
<label>Printer Key (sesuai config.json)</label>
<input type="text" id="printerKey" value="etiket">
</div>
</div>
<div class="actions">
<button class="primary" onclick="testCetak()">Cetak Sekarang</button>
</div>
</div>

<div class="card">
<h2>Output</h2>
<div class="output" id="output">(klik tombol di atas untuk lihat response)</div>
</div>

<div class="card">
<h2>Sample Integrasi — Laravel 12 + Livewire 4</h2>

<p style="font-size:0.85rem;margin-bottom:0.5rem;color:#475569;"><strong>1. Reusable Blade component</strong> <code>resources/views/components/print-button.blade.php</code></p>
<pre>@props(['printerKey', 'pdfPath', 'label' =&gt; 'Cetak'])

&lt;x-primary-button type="button"
    x-data
    x-on:click="cetakPDF(@js($pdfPath), @js($printerKey))"
    {{ $attributes }}&gt;
    {{ $label }}
&lt;/x-primary-button&gt;</pre>

<p style="font-size:0.85rem;margin-bottom:0.5rem;margin-top:0.75rem;color:#475569;"><strong>2. Pakai di halaman Livewire SFC</strong> (mis. <code>resources/views/pages/.../cetak-etiket.blade.php</code>)</p>
<pre>&lt;x-print-button printer-key="etiket"
                pdf-path="\\&lt;server&gt;\&lt;share&gt;\etiket.pdf"
                label="Cetak Etiket Apotek" /&gt;

&lt;x-print-button printer-key="kasir"
                pdf-path="C:\temp\struk.pdf"
                label="Cetak Struk" /&gt;</pre>

<p style="font-size:0.85rem;margin-bottom:0.5rem;margin-top:0.75rem;color:#475569;"><strong>3. Helper JS</strong> di <code>resources/js/app.js</code> (sekali registrasi global)</p>
<pre>window.cetakPDF = async function (pdfPath, printerKey) {
    try {
        const res = await fetch('http://localhost:%d/cetak-pdf', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ pathPdf: pdfPath, printerKey: printerKey }),
        });
        const data = await res.json();
        if (data.ok) {
            Livewire.dispatch('toast', { type: 'success', message: data.msg });
        } else {
            Livewire.dispatch('toast', { type: 'error', message: data.msg });
        }
    } catch (e) {
        Livewire.dispatch('toast', { type: 'error',
            message: 'Print agent tidak aktif di PC ini. Hubungi IT.' });
    }
};</pre>

<p style="font-size:0.85rem;margin-bottom:0.5rem;margin-top:0.75rem;color:#475569;"><strong>4. Generate PDF dulu, baru cetak</strong> (Livewire dispatch event ke browser)</p>
<pre>// Di SFC component (PHP)
public function generateDanCetak(int $rjNo): void
{
    $pdf = Pdf::loadView('pages.cetak.etiket-apotek', ['rjNo' =&gt; $rjNo]);
    $filename = Carbon::now()-&gt;format('dmYHis') . '.pdf';
    $path = 'upload/etiket/' . $filename;
    Storage::disk('local')-&gt;put($path, $pdf-&gt;output());

    // Path UNC kalau file di share, atau path lokal kalau di-mount
    $pcPath = '\\\\fileserver\\etiket\\' . $filename;

    $this-&gt;dispatch('print-pdf', pathPdf: $pcPath, printerKey: 'etiket');
}

// Listener di blade (sisi browser)
&lt;script&gt;
Livewire.on('print-pdf', ({ pathPdf, printerKey }) =&gt; {
    window.cetakPDF(pathPdf, printerKey);
});
&lt;/script&gt;</pre>
</div>

</div>

<script>
async function callApi(method, path, body) {
    const out = document.getElementById('output');
    out.textContent = method + ' ' + path + ' ...';
    try {
        const opts = { method };
        if (body) {
            opts.headers = { 'Content-Type': 'application/json' };
            opts.body = JSON.stringify(body);
        }
        const res = await fetch(path, opts);
        const data = await res.json();
        out.textContent = '[' + res.status + '] ' + method + ' ' + path + '\n\n' + JSON.stringify(data, null, 2);
    } catch (e) {
        out.textContent = 'ERROR: ' + e.message;
    }
}
function testCetak() {
    const pathPdf = document.getElementById('pathPdf').value.trim();
    const printerKey = document.getElementById('printerKey').value.trim();
    if (!pathPdf || !printerKey) { alert('Path PDF dan Printer Key wajib diisi'); return; }
    callApi('POST', '/cetak-pdf', { pathPdf, printerKey });
}
</script>
</body></html>
`

// GET /health — liveness check
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"agent":   agentName,
		"version": agentVersion,
		"time":    time.Now().Format(time.RFC3339),
	})
}

// GET /printers — daftar logical key printer yang terdaftar di config
func handlePrinters(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		keys := make([]string, 0, len(config.Printers))
		for k := range config.Printers {
			keys = append(keys, k)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":       true,
			"printers": config.Printers,
			"keys":     keys,
		})
	}
}

// GET /system-printers — list printer Windows asli (untuk debug / setup awal)
func handleSystemPrinters(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("wmic", "printer", "get", "name")
	output, err := cmd.Output()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "wmic gagal: "+err.Error())
		return
	}

	lines := strings.Split(string(output), "\n")
	var printers []string
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if i == 0 || line == "" {
			continue
		}
		printers = append(printers, line)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"printers": printers,
	})
}

// POST /cetak-pdf — translasi PROCEDURE CETAK_PDF_HOST dari Oracle Forms
//
// Body JSON:
//   { "pathPdf": "C:\\path\\to\\file.pdf", "printerKey": "etiket" }
//
// pathPdf bisa juga UNC path ke share: "\\\\172.8.8.12\\rad_path\\foto\\xxx.pdf"
func handleCetakPDF(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "POST only")
			return
		}

		var req CetakRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		// Validasi printerKey (mandatory)
		if strings.TrimSpace(req.PrinterKey) == "" {
			respondError(w, http.StatusBadRequest, "printerKey tidak boleh kosong")
			return
		}

		// Resolve printer key → printer name Windows
		printerName, ok := config.Printers[req.PrinterKey]
		if !ok {
			respondError(w, http.StatusBadRequest,
				fmt.Sprintf("printerKey '%s' tidak terdaftar di config", req.PrinterKey))
			return
		}

		// Tentukan path file untuk dicetak. Dua mode:
		//   1. pdfBase64 → decode → simpan temp file → print
		//   2. pathPdf → langsung pakai file path (UNC / lokal)
		var pathToPrint string
		var cleanupTemp string // temp file untuk dihapus setelah print

		if strings.TrimSpace(req.PdfBase64) != "" {
			// Mode 1: terima PDF as base64 (tidak butuh share/UNC, paling cocok untuk web app)
			data, err := base64.StdEncoding.DecodeString(req.PdfBase64)
			if err != nil {
				respondError(w, http.StatusBadRequest, "pdfBase64 invalid: "+err.Error())
				return
			}

			// Naming temp file: pakai filename dari request (sanitized), atau timestamp
			fname := strings.TrimSpace(req.Filename)
			if fname == "" {
				fname = "sirus-print-" + time.Now().Format("20060102-150405") + ".pdf"
			} else {
				fname = filepath.Base(fname) // anti path-traversal
				if !strings.HasSuffix(strings.ToLower(fname), ".pdf") {
					fname += ".pdf"
				}
			}

			tmpFile := filepath.Join(os.TempDir(), fname)
			if err := os.WriteFile(tmpFile, data, 0644); err != nil {
				respondError(w, http.StatusInternalServerError, "tulis temp file gagal: "+err.Error())
				return
			}
			pathToPrint = tmpFile
			cleanupTemp = tmpFile

		} else if strings.TrimSpace(req.PathPDF) != "" {
			// Mode 2: file already on disk (UNC share / lokal PC user)
			if _, err := os.Stat(req.PathPDF); os.IsNotExist(err) {
				respondError(w, http.StatusNotFound, "File tidak ditemukan: "+req.PathPDF)
				return
			}
			pathToPrint = req.PathPDF

		} else {
			respondError(w, http.StatusBadRequest,
				"Harus salah satu: pathPdf (path file) atau pdfBase64 (PDF content)")
			return
		}

		// Build & execute print command
		cmd, err := buildPrintCommand(config, pathToPrint, printerName)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			if cleanupTemp != "" {
				os.Remove(cleanupTemp)
			}
			return
		}

		log.Printf("[CETAK] %s → %s | mode: %s",
			pathToPrint, printerName,
			func() string {
				if cleanupTemp != "" {
					return "base64"
				}
				return "path"
			}())

		if err := cmd.Run(); err != nil {
			respondError(w, http.StatusInternalServerError,
				fmt.Sprintf("exec gagal: %v", err))
			if cleanupTemp != "" {
				os.Remove(cleanupTemp)
			}
			return
		}

		// Cleanup temp file setelah print sukses (spooler sudah copy file ke queue)
		if cleanupTemp != "" {
			// Delay sebentar supaya spooler sempat baca file (jaga-jaga kalau async)
			go func(p string) {
				time.Sleep(3 * time.Second)
				os.Remove(p)
			}(cleanupTemp)
		}

		writeJSON(w, http.StatusOK, APIResponse{
			OK:  true,
			Msg: fmt.Sprintf("Print terkirim → %s", printerName),
		})
	}
}

/* ===============================
 | HELPERS
 =============================== */

func respondError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, APIResponse{OK: false, Msg: msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

/* ===============================
 | MIDDLEWARE
 =============================== */

func corsMiddleware(allowed []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, a := range allowed {
			if a == "*" || a == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s (%v)", r.RemoteAddr, r.Method, r.URL.Path, time.Since(start))
	})
}
