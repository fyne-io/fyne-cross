package icon

import (
	"fmt"
	"image"
	"os"

	ico "github.com/Kodeworks/golang-image-ico"
)

const (
	// Default represents the default's icon name
	Default = "Icon.png"
)

// ConvertPngToIco convert a png file to ico format
func ConvertPngToIco(pngPath string, icoPath string) error {
	// convert icon
	img, err := os.Open(pngPath)
	if err != nil {
		return fmt.Errorf("failed to open source image: %s", err)
	}
	defer img.Close()
	srcImg, _, err := image.Decode(img)
	if err != nil {
		return fmt.Errorf("failed to decode source image: %s", err)
	}

	file, err := os.Create(icoPath)
	if err != nil {
		return fmt.Errorf("failed to open image file: %s", err)
	}
	defer file.Close()
	err = ico.Encode(file, srcImg)
	if err != nil {
		return fmt.Errorf("failed to write image file: %s", err)
	}
	return nil
}
