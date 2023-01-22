package converter

import (
	"image/jpeg"
	"image/png"
	"io"
)

func ConvertPNGToJPEG(w io.Writer, r io.Reader) error {
	img, err := png.Decode(r)
	if err != nil {
		return err
	}

	return jpeg.Encode(w, img, &jpeg.Options{Quality: 80})
}
