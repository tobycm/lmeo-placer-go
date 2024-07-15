package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	placePngUrl = "https://foloplace.tobycm.dev/place.png"
	wsUrl       = "wss://foloplace.tobycm.dev/ws"
)

var (
	canvas = Canvas{Mutex: &sync.Mutex{}}
	place  = Canvas{Mutex: &sync.Mutex{}}
	works  = Works{Queue: make([]*Work, 0), Mutex: &sync.Mutex{}}
)

func main() {
	imagePath := "./elysia.png"
	offset := [2]int{800, 800} // starting point, [x, y]

	if len(os.Args) > 3 {
		var err error
		offset, imagePath, err = parseArgs(os.Args)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	canvasImage, err := GetCanvasImage(placePngUrl)
	if err != nil {
		fmt.Println(err)
		return
	}
	canvas.FromImage(canvasImage)
	fmt.Println("Successfully fetched place.png")

	placeImage, err := loadImage(imagePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	place.FromImage(placeImage)
	fmt.Println("Successfully loaded image to place")

	for y := 0; y < place.Height; y++ {
		for x := 0; x < place.Width; x++ {
			cx, cy := x+offset[0], y+offset[1]

			r, g, b := place.At(x, y)
			cr, cg, cb := canvas.At(cx, cy)

			if r != cr || g != cg || b != cb {
				// fmt.Printf("Mismatch at %d, %d\n", cx, cy)
				works.Add(&Work{x: cx, y: cy, r: r, g: g, b: b})
			}
		}
	}

	masterWs := PlaceWs{Url: wsUrl}
	if err := masterWs.Connect(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Master websocket successfully connected")

	go func() {
		for {
			// time.Sleep(100 * time.Microsecond)
			_, message, err := masterWs.Conn.ReadMessage()
			if err != nil {
				fmt.Println(err)
				masterWs.Reconnect()
				return
			}

			// fmt.Println("Received message:", message)

			for len(message) >= 11 {
				x := binary.BigEndian.Uint32(message[0:4])
				y := binary.BigEndian.Uint32(message[4:8])

				if x < uint32(offset[0]) || y < uint32(offset[1]) || x >= uint32(offset[0]+place.Width) || y >= uint32(offset[1]+place.Height) {
					message = message[11:]
					time.Sleep(5 * time.Microsecond)
					continue
				}

				// fmt.Printf("Placed pixel at %d, %d\n", x, y)
				r := message[8]
				g := message[9]
				b := message[10]

				canvas.Set(int(x), int(y), r, g, b)

				pr, pg, pb := place.At(int(x)-offset[0], int(y)-offset[1])

				if r != pr || g != pg || b != pb {
					fmt.Printf("Mismatch at %d, %d\n", x, y)
					works.Add(&Work{x: int(x), y: int(y), r: pr, g: pg, b: pb})
				}

				message = message[11:]
			}
		}
	}()

	// time.Sleep(5 * time.Second)

	workers := 1000

	for i := 0; i < workers; i++ {
		go worker(i, &works)
	}

	fmt.Println("Works:", len(works.Queue))

	// go func() {
	// 	for {
	// 		time.Sleep(5 * time.Second)

	// 		fmt.Println("Works:", len(works.Queue))
	// 	}
	// }()

	// for {
	// 	time.Sleep(100 * time.Millisecond)

	// 	for y := 0; y < place.Height; y++ {
	// 		for x := 0; x < place.Width; x++ {
	// 			cx, cy := x+offset[0], y+offset[1]

	// 			r, g, b := place.At(x, y)
	// 			cr, cg, cb := canvas.At(cx, cy)

	// 			if r != cr || g != cg || b != cb {
	// 				// fmt.Printf("Mismatch at %d, %d\n", cx, cy)
	// 				works.Add(&Work{x: cx, y: cy, r: r, g: g, b: b})
	// 			}
	// 		}
	// 	}
	// }
}
