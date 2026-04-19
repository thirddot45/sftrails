package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func mustFace(ttf []byte, size float64) font.Face {
	f, err := opentype.Parse(ttf)
	if err != nil {
		log.Fatal(err)
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		log.Fatal(err)
	}
	return face
}

func drawText(img *image.RGBA, face font.Face, x, baseline int, c color.Color, s string) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.P(x, baseline),
	}
	d.DrawString(s)
}

func main() {
	const w, h = 1200, 630
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	bg := color.RGBA{0x11, 0x18, 0x27, 0xff}        // gray-900
	accent := color.RGBA{0x22, 0xc5, 0x5e, 0xff}    // green-500
	subtitle := color.RGBA{0xd1, 0xd5, 0xdb, 0xff}  // gray-300
	tagline := color.RGBA{0x9c, 0xa3, 0xaf, 0xff}   // gray-400
	smallText := color.RGBA{0x6b, 0x72, 0x80, 0xff} // gray-500

	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	// Bottom accent bar
	draw.Draw(img, image.Rect(0, h-24, w, h), &image.Uniform{accent}, image.Point{}, draw.Src)

	// Decorative "rideable / mixed / closed" stripe trio just under the title
	stripeY := 270
	draw.Draw(img, image.Rect(80, stripeY, 240, stripeY+10), &image.Uniform{accent}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(252, stripeY, 360, stripeY+10), &image.Uniform{color.RGBA{0xf5, 0x9e, 0x0b, 0xff}}, image.Point{}, draw.Src)
	draw.Draw(img, image.Rect(372, stripeY, 480, stripeY+10), &image.Uniform{color.RGBA{0xef, 0x44, 0x44, 0xff}}, image.Point{}, draw.Src)

	titleFace := mustFace(gobold.TTF, 120)
	subFace := mustFace(goregular.TTF, 44)
	tagFace := mustFace(goregular.TTF, 32)
	domainFace := mustFace(goregular.TTF, 28)

	drawText(img, titleFace, 80, 220, color.White, "SF Trails")
	drawText(img, subFace, 80, 360, subtitle, "South Florida Mountain Bike")
	drawText(img, subFace, 80, 414, subtitle, "Trail Status")
	drawText(img, tagFace, 80, 490, tagline, "Real-time, community-reported")
	drawText(img, tagFace, 80, 532, tagline, "trail conditions")
	drawText(img, domainFace, 80, 588, smallText, "sftrails.info")

	out := "static/og-image.png"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}
	f, err := os.Create(out)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}
