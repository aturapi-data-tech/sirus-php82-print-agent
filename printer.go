package main

import (
	"fmt"
	"os/exec"
)

// buildPrintCommand membuat *exec.Cmd berdasarkan jenis PDF reader
// yang ter-config (sumatra / foxit / pdftoprinter).
//
// Sintaks per reader:
//   sumatra      : SumatraPDF.exe -print-to "Printer" -silent file.pdf
//   foxit        : FoxitReader.exe /t "file.pdf" "Printer"
//   pdftoprinter : PDFtoPrinter.exe "file.pdf" "Printer"
func buildPrintCommand(config *Config, pathPdf, printerName string) (*exec.Cmd, error) {
	if config.PdfReaderPath == "" {
		return nil, fmt.Errorf("config.pdfReaderPath kosong")
	}

	switch config.PdfReaderType {
	case "sumatra", "":
		return exec.Command(
			config.PdfReaderPath,
			"-print-to", printerName,
			"-silent",
			"-exit-when-done",
			pathPdf,
		), nil

	case "foxit":
		return exec.Command(
			config.PdfReaderPath,
			"/t", pathPdf, printerName,
		), nil

	case "pdftoprinter":
		return exec.Command(
			config.PdfReaderPath,
			pathPdf, printerName,
		), nil

	default:
		return nil, fmt.Errorf("pdfReaderType tidak dikenal: %s", config.PdfReaderType)
	}
}
