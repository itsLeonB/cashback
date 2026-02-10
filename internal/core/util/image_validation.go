package util

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"io"
	"math"
	"net/http"

	_ "image/jpeg"
	_ "image/png"

	"github.com/itsLeonB/ungerr"
	_ "golang.org/x/image/webp"
)

// Validation constants
const (
	MinFileSize    = 50 * 1024 // 50 KB
	MinImageWidth  = 800
	MinImageHeight = 600
	BlurThreshold  = 50.0 // Conservative threshold for extreme blur
	MinBrightness  = 20.0
	MaxBrightness  = 235.0
	MinContrast    = 20.0
)

var (
	ErrImageTooSmall     = ungerr.UnprocessableEntityError("image file size is too small (minimum 50KB)")
	ErrInvalidImageType  = ungerr.UnprocessableEntityError("invalid image type (allowed: jpeg, png, webp)")
	ErrImageDecodeFailed = ungerr.UnprocessableEntityError("failed to decode image")
	ErrImageDimensions   = ungerr.UnprocessableEntityError("image dimensions too small (minimum 800x600)")
	ErrImageTooBlurry    = ungerr.UnprocessableEntityError("image is too blurry to read")
	ErrImageTooDark      = ungerr.UnprocessableEntityError("image is too dark")
	ErrImageTooBright    = ungerr.UnprocessableEntityError("image is too bright")
	ErrImageLowContrast  = ungerr.UnprocessableEntityError("image has too low contrast")
)

// ValidateImage performs strict validation on the image data.
// It checks file size, type, dimensions, blurriness, and exposure.
func ValidateImage(r io.Reader, fileSize int64) error {
	if fileSize < MinFileSize {
		return ErrImageTooSmall
	}

	// Read the entire image content into memory for processing
	// We need random access for decoding and processing
	data, err := io.ReadAll(r)
	if err != nil {
		return ungerr.Wrap(err, "failed to read image data")
	}

	// 1. File Type Validation
	if len(data) < 512 {
		return ErrImageDecodeFailed
	}
	contentType := http.DetectContentType(data[:512])
	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
		// Allowed
	default:
		return ErrInvalidImageType
	}

	// 2. Decode Image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		if errors.Is(err, image.ErrFormat) {
			return ErrInvalidImageType
		}
		return ungerr.Wrap(err, "failed to decode image")
	}

	// 3. Dimension Validation
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width < MinImageWidth || height < MinImageHeight {
		return ErrImageDimensions
	}

	// Convert to grayscale for analysis
	grayImg := toGrayscale(img)

	// 4. Blur Detection (Laplacian Variance)
	variance := laplacianVariance(grayImg)
	if variance < BlurThreshold {
		return ErrImageTooBlurry
	}

	// 5. Brightness & Contrast Validation
	brightness, contrast := calculateBrightnessAndContrast(grayImg)
	if brightness < MinBrightness {
		return ErrImageTooDark
	}
	if brightness > MaxBrightness {
		return ErrImageTooBright
	}
	if contrast < MinContrast {
		return ErrImageLowContrast
	}

	return nil
}

// toGrayscale converts an image to grayscale.
func toGrayscale(img image.Image) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			gray.Set(x, y, color.GrayModel.Convert(img.At(x, y)))
		}
	}
	return gray
}

// laplacianVariance computes the variance of the Laplacian of the image.
// This is a standard measure for image blurriness.
func laplacianVariance(img *image.Gray) float64 {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Laplacian kernel
	//  0  1  0
	//  1 -4  1
	//  0  1  0
	kernel := [3][3]int{
		{0, 1, 0},
		{1, -4, 1},
		{0, 1, 0},
	}

	var sum float64
	var sqSum float64
	count := 0

	// Convolve (ignore borders for simplicity)
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			var pixelVal int
			for ky := range 3 {
				for kx := range 3 {
					// Access pixel (x + kx - 1, y + ky - 1)
					// img.GrayAt is relative to Rect.Min, need to adjust if bounds don't start at 0,0
					// safely using At for simplicity or optimize with GrayAt
					p := img.GrayAt(bounds.Min.X+x+kx-1, bounds.Min.Y+y+ky-1).Y
					pixelVal += int(p) * kernel[ky][kx]
				}
			}
			val := float64(pixelVal)
			sum += val
			sqSum += val * val
			count++
		}
	}

	if count == 0 {
		return 0
	}

	mean := sum / float64(count)
	variance := (sqSum / float64(count)) - (mean * mean)
	return variance
}

// calculateBrightnessAndContrast computes the average brightness and standard deviation (contrast).
func calculateBrightnessAndContrast(img *image.Gray) (float64, float64) {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	count := width * height

	if count == 0 {
		return 0, 0
	}

	var sum float64
	var sqSum float64

	for y := range height {
		for x := range width {
			// GrayAt is optimized for image.Gray
			val := float64(img.GrayAt(bounds.Min.X+x, bounds.Min.Y+y).Y)
			sum += val
			sqSum += val * val
		}
	}

	mean := sum / float64(count)
	variance := (sqSum / float64(count)) - (mean * mean)
	stdDev := math.Sqrt(variance)

	return mean, stdDev
}
