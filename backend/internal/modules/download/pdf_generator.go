package download

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// PDFBuilder builds a clean, specification-compliant PDF 1.4 document.
type PDFBuilder struct {
	buf     bytes.Buffer
	offsets []int
}

func NewPDFBuilder() *PDFBuilder {
	b := &PDFBuilder{
		offsets: make([]int, 0, 10),
	}
	b.buf.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	return b
}

func (b *PDFBuilder) addObj(content string) int {
	objNum := len(b.offsets) + 1
	b.offsets = append(b.offsets, b.buf.Len())
	b.buf.WriteString(fmt.Sprintf("%d 0 obj\n%s\nendobj\n", objNum, strings.TrimSpace(content)))
	return objNum
}

// sanitizePDFText escapes special PDF string characters: \ ( )
func sanitizePDFText(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}

// wrapText splits a long text string into lines not exceeding maxChars.
func wrapText(text string, maxChars int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	var lines []string
	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= maxChars {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}

// GenerateOfficialDocumentPDF creates an authentic, styled official PDF document for Kabupaten Pidie downloads.
func GenerateOfficialDocumentPDF(title, category, description, filename string) []byte {
	b := NewPDFBuilder()

	// 1: Catalog -> 2 0 R
	b.addObj("<< /Type /Catalog /Pages 2 0 R >>")

	// 2: Pages -> Kids [3 0 R], Count 1
	b.addObj("<< /Type /Pages /Kids [3 0 R] /Count 1 >>")

	// Stream contents
	var stream bytes.Buffer

	// Top banner (Emerald green)
	stream.WriteString("q\n")
	stream.WriteString("0.024 0.478 0.345 rg\n") // #065f46
	stream.WriteString("0 826.89 595.28 15 re f\n")
	stream.WriteString("Q\n")

	// Header box
	stream.WriteString("q\n")
	stream.WriteString("0.96 0.98 0.96 rg\n")
	stream.WriteString("40 735 515.28 75 re f\n")
	stream.WriteString("0.1 0.45 0.32 RG 1.5 w\n")
	stream.WriteString("40 735 515.28 75 re S\n")
	stream.WriteString("Q\n")

	// Header Text
	stream.WriteString("BT\n")
	stream.WriteString("/F1 15 Tf\n")
	stream.WriteString("0.05 0.3 0.2 rg\n")
	stream.WriteString("55 785 Td (PEMERINTAH KABUPATEN PIDIE) Tj\n")
	stream.WriteString("ET\n")

	stream.WriteString("BT\n")
	stream.WriteString("/F2 10 Tf\n")
	stream.WriteString("0.35 0.4 0.45 rg\n")
	stream.WriteString("55 768 Td (Portal Informasi Resmi & Layanan Publik — pidiekab.web.id) Tj\n")
	stream.WriteString("ET\n")

	stream.WriteString("BT\n")
	stream.WriteString("/F2 9 Tf\n")
	stream.WriteString("0.4 0.45 0.5 rg\n")
	stream.WriteString(fmt.Sprintf("55 748 Td (Kategori Dokumen: %s | Terbit: %s) Tj\n", sanitizePDFText(category), time.Now().Format("02 January 2006")))
	stream.WriteString("ET\n")

	// Document Title Section
	stream.WriteString("BT\n")
	stream.WriteString("/F1 14 Tf\n")
	stream.WriteString("0.1 0.15 0.2 rg\n")
	stream.WriteString("40 695 Td\n")
	titleLines := wrapText(title, 60)
	for i, line := range titleLines {
		if i > 0 {
			stream.WriteString("0 -18 Td\n")
		}
		stream.WriteString(fmt.Sprintf("(%s) Tj\n", sanitizePDFText(line)))
	}
	stream.WriteString("ET\n")

	// Divider
	yCursor := 695 - float64(len(titleLines))*18 - 15
	stream.WriteString("q\n")
	stream.WriteString("0.8 0.85 0.9 RG 1 w\n")
	stream.WriteString(fmt.Sprintf("40 %.2f m 555.28 %.2f l S\n", yCursor, yCursor))
	stream.WriteString("Q\n")

	yCursor -= 25

	// Description Section
	stream.WriteString("BT\n")
	stream.WriteString("/F1 11 Tf\n")
	stream.WriteString("0.05 0.3 0.2 rg\n")
	stream.WriteString(fmt.Sprintf("40 %.2f Td (I. DESKRIPSI & RINGKASAN DOKUMEN) Tj\n", yCursor))
	stream.WriteString("ET\n")

	yCursor -= 18

	stream.WriteString("BT\n")
	stream.WriteString("/F2 10 Tf\n")
	stream.WriteString("0.2 0.25 0.3 rg\n")
	stream.WriteString(fmt.Sprintf("40 %.2f Td\n", yCursor))

	descLines := wrapText(description, 75)
	for i, line := range descLines {
		if i > 0 {
			stream.WriteString("0 -15 Td\n")
		}
		stream.WriteString(fmt.Sprintf("(%s) Tj\n", sanitizePDFText(line)))
	}
	stream.WriteString("ET\n")

	yCursor -= float64(len(descLines))*15 + 25

	// Detail Section (Authentic Pidie Information)
	stream.WriteString("BT\n")
	stream.WriteString("/F1 11 Tf\n")
	stream.WriteString("0.05 0.3 0.2 rg\n")
	stream.WriteString(fmt.Sprintf("40 %.2f Td (II. INFORMASI WILAYAH & REKOMENDASI RESMI) Tj\n", yCursor))
	stream.WriteString("ET\n")

	yCursor -= 18

	infoLines := []string{
		"Kabupaten Pidie merupakan salah satu lumbung pangan dan pusat kebudayaan",
		"bersejarah di Provinsi Aceh dengan 23 kecamatan dan kekayaan alam yang melimpah.",
		"",
		"Destinasi & Potensi Utama:",
		"- Wisata Alam & Geopark: Tangse (Lembah Sejuk & Kopi Robusta), Guha Tujoh Laweung.",
		"- Wisata Sejarah & Religi: Masjid Po Teumeureuhom, Kompleks Makam Daud Beureueh.",
		"- Kuliner Khas: Emping Melinjo Asli Pidie, Kupi Khop (Kopi Terbalik Sigli), Meukutop.",
		"- Potensi Ekonomi: Pertanian Padi Organik, Perkebunan Kopi, Sektor Kelautan & Perikanan.",
	}

	stream.WriteString("BT\n")
	stream.WriteString("/F2 10 Tf\n")
	stream.WriteString("0.2 0.25 0.3 rg\n")
	stream.WriteString(fmt.Sprintf("40 %.2f Td\n", yCursor))
	for i, line := range infoLines {
		if i > 0 {
			stream.WriteString("0 -15 Td\n")
		}
		if strings.HasPrefix(line, "-") || strings.HasSuffix(line, ":") {
			stream.WriteString(fmt.Sprintf("(%s) Tj\n", sanitizePDFText(line)))
		} else if line == "" {
			stream.WriteString("() Tj\n")
		} else {
			stream.WriteString(fmt.Sprintf("(%s) Tj\n", sanitizePDFText(line)))
		}
	}
	stream.WriteString("ET\n")

	yCursor -= float64(len(infoLines))*15 + 25

	// Metadata Box
	stream.WriteString("q\n")
	stream.WriteString("0.95 0.96 0.98 rg\n")
	stream.WriteString(fmt.Sprintf("40 %.2f 515.28 65 re f\n", yCursor-65))
	stream.WriteString("0.8 0.85 0.9 RG 1 w\n")
	stream.WriteString(fmt.Sprintf("40 %.2f 515.28 65 re S\n", yCursor-65))
	stream.WriteString("Q\n")

	stream.WriteString("BT\n")
	stream.WriteString("/F1 9 Tf\n")
	stream.WriteString("0.1 0.2 0.35 rg\n")
	stream.WriteString(fmt.Sprintf("55 %.2f Td (LEMBAR VERIFIKASI DIGITAL:) Tj\n", yCursor-18))
	stream.WriteString("ET\n")

	stream.WriteString("BT\n")
	stream.WriteString("/F2 8.5 Tf\n")
	stream.WriteString("0.3 0.35 0.4 rg\n")
	stream.WriteString(fmt.Sprintf("55 %.2f Td (Nama Berkas: %s | Format: PDF 1.4 Standar Internasional) Tj\n", yCursor-32, sanitizePDFText(filename)))
	stream.WriteString("0 -14 Td\n")
	stream.WriteString("(Dikeluarkan Oleh: Pemerintah Kabupaten Pidie, Provinsi Aceh, Indonesia) Tj\n")
	stream.WriteString("0 -14 Td\n")
	stream.WriteString("(Akses Publik Resmi: https://pidiekab.web.id/id/unduhan/) Tj\n")
	stream.WriteString("ET\n")

	// Footer Bar
	stream.WriteString("q\n")
	stream.WriteString("0.85 0.88 0.9 RG 1 w\n")
	stream.WriteString("40 50 m 555.28 50 l S\n")
	stream.WriteString("Q\n")

	stream.WriteString("BT\n")
	stream.WriteString("/F2 8 Tf\n")
	stream.WriteString("0.5 0.55 0.6 rg\n")
	stream.WriteString("40 38 Td (Hak Cipta Pemerintah Kabupaten Pidie — Sigli, Aceh. Dokumen ini dapat diunduh bebas untuk keperluan publik.) Tj\n")
	stream.WriteString("ET\n")

	stream.WriteString("q\n")
	stream.WriteString("0.024 0.478 0.345 rg\n")
	stream.WriteString("0 0 595.28 6 re f\n")
	stream.WriteString("Q\n")

	streamBytes := stream.Bytes()

	// 3: Page -> Parent 2, Contents 4, Resources (Fonts 5, 6)
	b.addObj("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595.28 841.89] /Contents 4 0 R /Resources << /Font << /F1 5 0 R /F2 6 0 R >> >> >>")

	// 4: Stream object
	b.addObj(fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(streamBytes), streamBytes))

	// 5: Font F1 (Helvetica-Bold)
	b.addObj("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold >>")

	// 6: Font F2 (Helvetica)
	b.addObj("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")

	// Cross-reference table
	xrefOffset := b.buf.Len()
	b.buf.WriteString(fmt.Sprintf("xref\n0 %d\n", len(b.offsets)+1))
	b.buf.WriteString("0000000000 65535 f \n")
	for _, offset := range b.offsets {
		b.buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
	}

	// Trailer
	b.buf.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\n", len(b.offsets)+1))
	b.buf.WriteString(fmt.Sprintf("startxref\n%d\n%%%%EOF\n", xrefOffset))

	return b.buf.Bytes()
}
