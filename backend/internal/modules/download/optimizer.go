package download

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"path/filepath"
	"strings"
)

const (
	// MaxDownloadFileSizeBytes enforces a strict 10 MB maximum limit per uploaded file.
	MaxDownloadFileSizeBytes = 10 * 1024 * 1024

	// MaxImageDimension limits image dimensions for brochure/infographic/map uploads to 1920px.
	MaxImageDimension = 1920

	// JpegOptimalQuality provides near-lossless clarity at minimal file size.
	JpegOptimalQuality = 82
)

var (
	ErrFileTooLarge       = errors.New("ukuran berkas melebihi batas maksimum 10 MB")
	ErrDisallowedExtension = errors.New("tipe berkas tidak diizinkan. Hanya berkas dokumen (PDF, Word, Excel, PowerPoint, ZIP, Gambar, CSV, TXT) yang diizinkan")
	ErrEmptyFile          = errors.New("berkas kosong tidak dapat diunggah")
)

var allowedExtensions = map[string]string{
	// Documents
	"pdf":  "PDF",
	"doc":  "DOC",
	"docx": "DOCX",
	"xls":  "XLS",
	"xlsx": "XLSX",
	"ppt":  "PPT",
	"pptx": "PPTX",
	"odt":  "ODT",
	"ods":  "ODS",
	"odp":  "ODP",
	"rtf":  "RTF",

	// Archives
	"zip": "ZIP",
	"rar": "RAR",
	"7z":  "7Z",

	// Text & Data
	"csv":  "CSV",
	"txt":  "TXT",
	"json": "JSON",
	"xml":  "XML",

	// Images / Maps / Infographics
	"jpg":  "JPG",
	"jpeg": "JPEG",
	"png":  "PNG",
	"webp": "WEBP",
	"svg":  "SVG",
}

// IsAllowedExtension checks if a given extension is within safe document types.
func IsAllowedExtension(ext string) (string, bool) {
	clean := strings.ToLower(strings.TrimPrefix(ext, "."))
	fileType, ok := allowedExtensions[clean]
	return fileType, ok
}

// OptimizeDownloadFile validates size limits and compresses/optimizes data before storage.
func OptimizeDownloadFile(filename string, data []byte) ([]byte, string, int64, error) {
	if len(data) == 0 {
		return nil, "", 0, ErrEmptyFile
	}

	if len(data) > MaxDownloadFileSizeBytes {
		return nil, "", 0, fmt.Errorf("%w (%d bytes > %d bytes)", ErrFileTooLarge, len(data), MaxDownloadFileSizeBytes)
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	fileType, allowed := IsAllowedExtension(ext)
	if !allowed {
		return nil, "", 0, ErrDisallowedExtension
	}

	// 1. Image documents (maps, infographics, brochures): compress and downscale if oversized
	switch ext {
	case "jpg", "jpeg", "png":
		optimized := compressImage(data, ext)
		if len(optimized) > 0 && len(optimized) < len(data) {
			return optimized, fileType, int64(len(optimized)), nil
		}
		return data, fileType, int64(len(data)), nil

	case "txt", "csv", "json", "xml", "svg":
		// Trim unnecessary trailing nulls and whitespace
		trimmed := bytes.TrimRight(data, "\r\n\t \x00")
		if len(trimmed) > 0 {
			return trimmed, fileType, int64(len(trimmed)), nil
		}
		return data, fileType, int64(len(data)), nil

	case "pdf":
		// Compress uncompressed PDF streams if present
		compressed := compressPDFStreams(data)
		if len(compressed) > 0 && len(compressed) < len(data) {
			return compressed, fileType, int64(len(compressed)), nil
		}
		return data, fileType, int64(len(data)), nil

	default:
		// Word, Excel, ZIP, etc. (already compressed archives)
		return data, fileType, int64(len(data)), nil
	}
}

// compressImage performs high-efficiency compression on image documents.
func compressImage(data []byte, format string) []byte {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data
	}

	b := img.Bounds()
	curW := b.Dx()
	curH := b.Dy()

	targetW := curW
	targetH := curH

	if curW > MaxImageDimension || curH > MaxImageDimension {
		ratio := math.Min(float64(MaxImageDimension)/float64(curW), float64(MaxImageDimension)/float64(curH))
		targetW = int(float64(curW) * ratio)
		targetH = int(float64(curH) * ratio)
		if targetW < 1 {
			targetW = 1
		}
		if targetH < 1 {
			targetH = 1
		}
		img = resizeImage(img, targetW, targetH)
	}

	var out bytes.Buffer
	if format == "jpg" || format == "jpeg" || (format == "png" && !imageHasAlpha(img)) {
		err = jpeg.Encode(&out, img, &jpeg.Options{Quality: JpegOptimalQuality})
	} else if format == "png" {
		enc := png.Encoder{CompressionLevel: png.BestCompression}
		err = enc.Encode(&out, img)
	}

	if err == nil && out.Len() > 0 && out.Len() < len(data) {
		return out.Bytes()
	}
	return data
}

func imageHasAlpha(img image.Image) bool {
	b := img.Bounds()
	step := 4
	for y := b.Min.Y; y < b.Max.Y; y += step {
		for x := b.Min.X; x < b.Max.X; x += step {
			_, _, _, a := img.At(x, y).RGBA()
			if a < 0xFFFF {
				return true
			}
		}
	}
	return false
}

func resizeImage(src image.Image, targetW, targetH int) image.Image {
	b := src.Bounds()
	srcW := b.Dx()
	srcH := b.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	xRatio := float64(srcW-1) / float64(targetW)
	yRatio := float64(srcH-1) / float64(targetH)

	for y := 0; y < targetH; y++ {
		sy := int(float64(y) * yRatio)
		for x := 0; x < targetW; x++ {
			sx := int(float64(x) * xRatio)
			dst.Set(x, y, src.At(b.Min.X+sx, b.Min.Y+sy))
		}
	}
	return dst
}

// compressPDFStreams deflates uncompressed streams in standard PDF documents when beneficial.
func compressPDFStreams(data []byte) []byte {
	// If file does not start with %PDF, do not alter
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return data
	}

	// For safety, if PDF is encrypted or linearized, preserve original
	if bytes.Contains(data, []byte("/Encrypt")) || bytes.Contains(data, []byte("/Linearized")) {
		return data
	}

	streamMarker := []byte("stream\n")
	streamMarkerCRLF := []byte("stream\r\n")
	endstreamMarker := []byte("endstream")

	var result bytes.Buffer
	result.Grow(len(data))

	offset := 0
	modified := false

	for {
		idx := bytes.Index(data[offset:], []byte("stream"))
		if idx == -1 {
			result.Write(data[offset:])
			break
		}

		streamPos := offset + idx
		// Check preceding dictionary for /Filter
		dictStart := bytes.LastIndex(data[:streamPos], []byte("<<"))
		if dictStart == -1 || dictStart < offset {
			result.Write(data[offset : streamPos+6])
			offset = streamPos + 6
			continue
		}

		dictBytes := data[dictStart:streamPos]
		// If already filtered, skip compression
		if bytes.Contains(dictBytes, []byte("/Filter")) {
			result.Write(data[offset : streamPos+6])
			offset = streamPos + 6
			continue
		}

		// Find end of "stream\n" or "stream\r\n"
		var contentStart int
		if bytes.HasPrefix(data[streamPos:], streamMarkerCRLF) {
			contentStart = streamPos + len(streamMarkerCRLF)
		} else if bytes.HasPrefix(data[streamPos:], streamMarker) {
			contentStart = streamPos + len(streamMarker)
		} else {
			result.Write(data[offset : streamPos+6])
			offset = streamPos + 6
			continue
		}

		// Find endstream
		endIdx := bytes.Index(data[contentStart:], endstreamMarker)
		if endIdx == -1 {
			result.Write(data[offset:])
			break
		}

		contentEnd := contentStart + endIdx
		// Trim optional trailing \r or \n before endstream
		actualContent := data[contentStart:contentEnd]
		actualContent = bytes.TrimRight(actualContent, "\r\n")

		// Only compress if content is non-trivial (> 64 bytes)
		if len(actualContent) > 64 {
			var deflated bytes.Buffer
			zw, err := zlib.NewWriterLevel(&deflated, zlib.BestCompression)
			if err == nil {
				_, _ = zw.Write(actualContent)
				_ = zw.Close()

				if deflated.Len() < len(actualContent) {
					// We can compress this stream!
					// Write up to dictStart
					result.Write(data[offset:dictStart])

					// Replace /Length in dictionary and add /Filter /FlateDecode
					dictStr := string(dictBytes)
					// Remove existing /Length ...
					dictStr = strings.TrimSpace(dictStr)
					if strings.HasSuffix(dictStr, ">>") {
						dictStr = strings.TrimSuffix(dictStr, ">>")
					}
					// Clean old /Length if present
					if lIdx := strings.Index(dictStr, "/Length"); lIdx != -1 {
						rest := dictStr[lIdx:]
						if endL := strings.IndexAny(rest[7:], " /\n\r>"); endL != -1 {
							dictStr = dictStr[:lIdx] + rest[7+endL:]
						}
					}
					newDict := fmt.Sprintf("%s /Length %d /Filter /FlateDecode >>\n", strings.TrimSpace(dictStr), deflated.Len())
					result.WriteString(newDict)

					// Write stream
					result.WriteString("stream\n")
					result.Write(deflated.Bytes())
					result.WriteString("\nendstream")

					offset = contentEnd + len(endstreamMarker)
					modified = true
					continue
				}
			}
		}

		result.Write(data[offset : contentEnd+len(endstreamMarker)])
		offset = contentEnd + len(endstreamMarker)
	}

	if modified && result.Len() < len(data) {
		return result.Bytes()
	}
	return data
}
