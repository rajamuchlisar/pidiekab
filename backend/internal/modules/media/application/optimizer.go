package application

import (
	"bytes"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"math"
)

const (
	// MaxDimension is the maximum width or height for uploaded images (1920px Full HD).
	MaxDimension = 1920
	// JpegQuality is the optimal quality trade-off (82) producing clean visual quality with minimal byte size.
	JpegQuality = 82
)

// OptimizeImage optimizes and compresses JPEG/PNG images down to max dimensions
// with high visual fidelity and optimal file size. If compression does not decrease file size
// or if decoding fails, it safely returns the original buffer and dimensions.
func OptimizeImage(data []byte, mimeType string, origW, origH int) ([]byte, int, int, string) {
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		return data, origW, origH, mimeType
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data, origW, origH, mimeType
	}

	b := img.Bounds()
	curW := b.Dx()
	curH := b.Dy()

	targetW := curW
	targetH := curH

	if curW > MaxDimension || curH > MaxDimension {
		ratio := math.Min(float64(MaxDimension)/float64(curW), float64(MaxDimension)/float64(curH))
		targetW = int(float64(curW) * ratio)
		targetH = int(float64(curH) * ratio)
		if targetW < 1 {
			targetW = 1
		}
		if targetH < 1 {
			targetH = 1
		}
		img = resizeBilinear(img, targetW, targetH)
	}

	var out bytes.Buffer
	var newMime string

	if format == "jpeg" || (format == "png" && !hasTransparency(img)) {
		err = jpeg.Encode(&out, img, &jpeg.Options{Quality: JpegQuality})
		newMime = "image/jpeg"
	} else if format == "png" {
		enc := png.Encoder{CompressionLevel: png.BestCompression}
		err = enc.Encode(&out, img)
		newMime = "image/png"
	}

	if err == nil && out.Len() > 0 && out.Len() < len(data) {
		return out.Bytes(), targetW, targetH, newMime
	}

	return data, curW, curH, mimeType
}

func hasTransparency(img image.Image) bool {
	b := img.Bounds()
	step := 1
	if b.Dx() > 100 || b.Dy() > 100 {
		step = 4
	}
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

// resizeBilinear resizes an image to targetW x targetH using high-quality bilinear interpolation.
func resizeBilinear(src image.Image, targetW, targetH int) image.Image {
	b := src.Bounds()
	srcW := b.Dx()
	srcH := b.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))

	xRatio := float64(srcW-1) / float64(targetW)
	yRatio := float64(srcH-1) / float64(targetH)

	for y := 0; y < targetH; y++ {
		srcY := float64(y) * yRatio
		yFloor := int(srcY)
		yDiff := srcY - float64(yFloor)
		yCeil := yFloor + 1
		if yCeil >= srcH {
			yCeil = yFloor
		}

		for x := 0; x < targetW; x++ {
			srcX := float64(x) * xRatio
			xFloor := int(srcX)
			xDiff := srcX - float64(xFloor)
			xCeil := xFloor + 1
			if xCeil >= srcW {
				xCeil = xFloor
			}

			r1, g1, b1, a1 := colorToRGBA(src.At(b.Min.X+xFloor, b.Min.Y+yFloor))
			r2, g2, b2, a2 := colorToRGBA(src.At(b.Min.X+xCeil, b.Min.Y+yFloor))
			r3, g3, b3, a3 := colorToRGBA(src.At(b.Min.X+xFloor, b.Min.Y+yCeil))
			r4, g4, b4, a4 := colorToRGBA(src.At(b.Min.X+xCeil, b.Min.Y+yCeil))

			w1 := (1 - xDiff) * (1 - yDiff)
			w2 := xDiff * (1 - yDiff)
			w3 := (1 - xDiff) * yDiff
			w4 := xDiff * yDiff

			red := uint8(math.Round(w1*r1 + w2*r2 + w3*r3 + w4*r4))
			green := uint8(math.Round(w1*g1 + w2*g2 + w3*g3 + w4*g4))
			blue := uint8(math.Round(w1*b1 + w2*b2 + w3*b3 + w4*b4))
			alpha := uint8(math.Round(w1*a1 + w2*a2 + w3*a3 + w4*a4))

			dst.SetRGBA(x, y, color.RGBA{R: red, G: green, B: blue, A: alpha})
		}
	}

	return dst
}

func colorToRGBA(c color.Color) (float64, float64, float64, float64) {
	r, g, b, a := c.RGBA()
	return float64(r >> 8), float64(g >> 8), float64(b >> 8), float64(a >> 8)
}
