package util_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/itsLeonB/cashback/internal/core/util"
	"github.com/stretchr/testify/assert"
)

func createSolidImage(width, height int, col color.Color) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			img.Set(x, y, col)
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func createNoiseImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			// Checkerboard pattern for high frequency/variance
			if (x+y)%2 == 0 {
				img.Set(x, y, color.White)
			} else {
				img.Set(x, y, color.Black)
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func createGradientImage(width, height int) []byte {
	img := image.NewGray(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			// Smooth horizontal gradient
			val := uint8((float64(x) / float64(width)) * 255)
			img.SetGray(x, y, color.Gray{Y: val})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

const ValidFileSize = 51 * 1024 // 51 KB

func TestValidateImageDimensions(t *testing.T) {
	// Too small
	smallImg := createSolidImage(100, 100, color.White)
	err := util.ValidateImage(bytes.NewReader(smallImg), ValidFileSize)
	if err != nil && err != util.ErrImageDimensions {
		t.Logf("Got error: %v, Expected: %v", err, util.ErrImageDimensions)
	}
	assert.Equal(t, util.ErrImageDimensions, err)

	// Valid size
	validImg := createSolidImage(800, 600, color.RGBA{128, 128, 128, 255})
	// Note: Solid image will fail blur check, but let's check it passes dimension check at least
	// Actually, ValidateImage returns first error.
	// 800x600 is min.

	// We expect ErrImageTooBlurry for solid image, which means it passed dimension check
	err = util.ValidateImage(bytes.NewReader(validImg), ValidFileSize)
	// Expect blur error, not dimension error
	assert.Equal(t, util.ErrImageTooBlurry, err)
}

func TestValidateImageFileSize(t *testing.T) {
	// Create a dummy reader with small size
	err := util.ValidateImage(bytes.NewReader([]byte("dummy")), 100)
	assert.Equal(t, util.ErrImageTooSmall, err)
}

func TestValidateImageBlur(t *testing.T) {
	// Solid color -> Very blurry (variance 0)
	solidImg := createSolidImage(800, 600, color.RGBA{128, 128, 128, 255})
	err := util.ValidateImage(bytes.NewReader(solidImg), ValidFileSize)
	if err != nil && err != util.ErrImageTooBlurry {
		t.Logf("Solid Image Error: %v", err)
	}
	assert.Equal(t, util.ErrImageTooBlurry, err)

	// Gradient -> Blurry (variance low) (Laplacian of linear gradient is 0)
	gradImg := createGradientImage(800, 600)
	err = util.ValidateImage(bytes.NewReader(gradImg), ValidFileSize)
	if err != nil && err != util.ErrImageTooBlurry {
		t.Logf("Gradient Image Error: %v", err)
	}
	assert.Equal(t, util.ErrImageTooBlurry, err)

	// Noise/Sharp edges -> Not blurry
	noiseImg := createNoiseImage(800, 600)
	err = util.ValidateImage(bytes.NewReader(noiseImg), ValidFileSize)
	// Noise image should pass blur check.
	// It might hit contrast check or brightness check depending on generation.
	// Black and White noise:
	// Mean brightness ~ 127.
	// Contrast ~ 127.
	// Should pass all.
	if err != nil {
		t.Logf("Noise Image Unexpected Error: %v", err)
	}
	assert.NoError(t, err)
}

func TestValidateImageBrightness(t *testing.T) {
	// Too dark
	// Use noise to pass blur check
	// add some noise to make it not blurry
	// manually inject some noise/edges?
	// It's hard to make a dark image that isn't solid without complex drawing.
	// However, Evaluate logic order:
	// Dimensions -> Blur -> Brightness.
	// So if we want to test Brightness, we must PASS Blur.
	// A dark noise image.

	img := image.NewGray(image.Rect(0, 0, 800, 600))
	for y := range 600 {
		for x := range 800 {
			if (x+y)%2 == 0 {
				img.SetGray(x, y, color.Gray{Y: 0})
			} else {
				img.SetGray(x, y, color.Gray{Y: 15}) // Averge ~7.5
			}
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		panic(err)
	}
	data := buf.Bytes()

	err := util.ValidateImage(bytes.NewReader(data), ValidFileSize)
	assert.Equal(t, util.ErrImageTooDark, err)

	// Too bright
	imgBright := image.NewGray(image.Rect(0, 0, 800, 600))
	for y := range 600 {
		for x := range 800 {
			if (x+y)%2 == 0 {
				imgBright.SetGray(x, y, color.Gray{Y: 240})
			} else {
				imgBright.SetGray(x, y, color.Gray{Y: 255}) // Average ~247.5
			}
		}
	}
	var buf2 bytes.Buffer
	err = jpeg.Encode(&buf2, imgBright, nil)
	assert.NoError(t, err)
	data2 := buf2.Bytes()

	err = util.ValidateImage(bytes.NewReader(data2), ValidFileSize)
	assert.Equal(t, util.ErrImageTooBright, err)
}

func TestValidateImageContrast(t *testing.T) {
	// Low contrast (mid-gray noise with low variance)
	// E.g. pixels 127 and 128.
	img := image.NewGray(image.Rect(0, 0, 800, 600))
	for y := range 600 {
		for x := range 800 {
			if (x+y)%2 == 0 {
				img.SetGray(x, y, color.Gray{Y: 127})
			} else {
				img.SetGray(x, y, color.Gray{Y: 130})
			}
		}
	}
	// Laplacian variance will be non-zero but small?
	// 127, 130 difference is 3.
	// Laplacian: 3 difference. Variance roughly proportional to square of diff.
	// It might trigger Blur check first if variance is < 50.
	// Let's ensure variance is high enough?
	// Checkerboard of 127 and 130.
	// Center 130 enclosed by 127s.
	// Laplacian: 127*4 - 130*4 ... wait.
	// Kernel: 0 1 0; 1 -4 1; 0 1 0.
	// Center 130. Neighbors 127.
	// 127*4 - 130*4 = -12.
	// Variance of (-12, 12) is 144. > 50. So Blur check passes.

	// Contrast (StdDev):
	// Values 127, 130. Mean 128.5.
	// Variance: ((127-128.5)^2 + (130-128.5)^2)/2 = (2.25 + 2.25)/2 = 2.25.
	// StdDev = sqrt(2.25) = 1.5.
	// 1.5 < MinContrast (20.0).
	// So should fail Contrast check.

	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	assert.NoError(t, err)
	data := buf.Bytes()

	err = util.ValidateImage(bytes.NewReader(data), ValidFileSize)
	if err != nil && err != util.ErrImageLowContrast {
		t.Logf("Contrast Error: %v", err)
	}
	assert.Equal(t, util.ErrImageLowContrast, err)
}
