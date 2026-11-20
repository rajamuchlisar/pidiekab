package download

import (
	"bytes"
	"strings"
	"testing"
)

func TestGenerateOfficialDocumentPDF(t *testing.T) {
	title := "Panduan Wisata & Kebudayaan Kabupaten Pidie"
	cat := "Panduan Wisata"
	desc := "Panduan lengkap destinasi wisata alam, sejarah, dan sentra kuliner khas di Kabupaten Pidie."
	filename := "panduan-wisata-pidie.pdf"

	pdfData := GenerateOfficialDocumentPDF(title, cat, desc, filename)

	if len(pdfData) == 0 {
		t.Fatal("expected non-empty PDF data")
	}

	// Must start with PDF header
	if !bytes.HasPrefix(pdfData, []byte("%PDF-1.4")) {
		t.Errorf("expected PDF-1.4 header, got %s", string(pdfData[:8]))
	}

	// Must end with %%EOF
	str := string(pdfData)
	if !strings.Contains(str, "%%EOF") {
		t.Error("expected end-of-file trailer marker")
	}

	// Must contain essential PDF elements
	for _, term := range []string{"/Catalog", "/Pages", "/Type1", "xref", "trailer"} {
		if !strings.Contains(str, term) {
			t.Errorf("missing expected PDF structural token: %s", term)
		}
	}
}
