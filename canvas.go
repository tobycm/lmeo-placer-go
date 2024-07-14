package main

import (
	"image"
	"sync"
)

type Canvas struct {
	Width  int
	Height int
	Data   []byte

	Mutex *sync.Mutex
}

func (canvas *Canvas) FromImage(img *image.Image) {
	canvas.Width = (*img).Bounds().Dx()
	canvas.Height = (*img).Bounds().Dy()

	canvas.Mutex.Lock()
	defer canvas.Mutex.Unlock()

	canvas.Data = make([]byte, canvas.Width*canvas.Height*3)

	for y := 0; y < canvas.Height; y++ {
		for x := 0; x < canvas.Width; x++ {
			r, g, b, _ := (*img).At(x, y).RGBA()
			r = r >> 8
			g = g >> 8
			b = b >> 8

			canvas.Data[(y*canvas.Width+x)*3] = uint8(r)
			canvas.Data[(y*canvas.Width+x)*3+1] = uint8(g)
			canvas.Data[(y*canvas.Width+x)*3+2] = uint8(b)
		}
	}
}

func (canvas *Canvas) At(x, y int) (uint8, uint8, uint8) {
	canvas.Mutex.Lock()
	defer canvas.Mutex.Unlock()

	return canvas.Data[(y*canvas.Width+x)*3], canvas.Data[(y*canvas.Width+x)*3+1], canvas.Data[(y*canvas.Width+x)*3+2]
}

func (canvas *Canvas) Set(x, y int, r, g, b uint8) {
	canvas.Mutex.Lock()
	defer canvas.Mutex.Unlock()

	canvas.Data[(y*canvas.Width+x)*3] = r
	canvas.Data[(y*canvas.Width+x)*3+1] = g
	canvas.Data[(y*canvas.Width+x)*3+2] = b
}
