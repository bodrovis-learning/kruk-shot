package capture

import "image"

// bgraToRGBA converts a BGRA pixel buffer into an RGBA image.
//
// The source may contain padding at the end of each row.
// The resulting image is always opaque.
func bgraToRGBA(
	pixels []byte,
	width int,
	height int,
	rowPitch int,
) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := range height {
		srcRow := pixels[y*rowPitch : y*rowPitch+width*bytesPerPixel]
		dstRow := img.Pix[y*img.Stride : y*img.Stride+width*bytesPerPixel]

		for x := range width {
			offset := x * bytesPerPixel

			// Swap blue and red channels.
			dstRow[offset] = srcRow[offset+2]
			dstRow[offset+1] = srcRow[offset+1]
			dstRow[offset+2] = srcRow[offset]

			// Desktop screenshots are opaque.
			dstRow[offset+3] = 255
		}
	}

	return img
}
