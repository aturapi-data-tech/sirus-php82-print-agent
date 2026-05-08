package main

import (
	"encoding/json"
	"os"
)

// Config = struktur config.json yang dibaca saat startup.
//
// Field:
//   Port            : port HTTP local agent (default 9999)
//   PdfReaderPath   : path absolute ke SumatraPDF.exe / pdf reader lain
//   PdfReaderType   : 'sumatra' | 'foxit' | 'pdftoprinter' — menentukan sintaks CLI
//   AllowedOrigins  : whitelist origin browser yang boleh akses agent (CORS)
//   Printers        : mapping logical-key (mis. "etiket") → Windows printer name
type Config struct {
	Port           int               `json:"port"`
	PdfReaderPath  string            `json:"pdfReaderPath"`
	PdfReaderType  string            `json:"pdfReaderType"`
	AllowedOrigins []string          `json:"allowedOrigins"`
	Printers       map[string]string `json:"printers"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if c.Port == 0 {
		c.Port = 9999
	}
	if c.PdfReaderType == "" {
		c.PdfReaderType = "sumatra"
	}
	return &c, nil
}
