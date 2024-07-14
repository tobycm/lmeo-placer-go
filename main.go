package main

import (
	"encoding/binary"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var placePngUrl = "https://foloplace.tobycm.dev/place.png"
var wsUrl = "wss://foloplace.tobycm.dev/ws"
var imagePath = "./elysia.png"

var offset = [2]int{800, 800} // starting point, [x, y]

func getPlacePng(url string) (*image.Image, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error getting place.png: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("received non 200 response code: %d", response.StatusCode)
	}

	img, _, err := image.Decode(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}

	return &img, nil
}

func connectWs(url string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("error connecting to websocket: %w", err)
	}

	return conn, nil
}

func loadImage(path string) (*image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %w", err)
	}

	return &img, nil
}

type Canvas struct {
	Width  int
	Height int
	Data   []byte

	Mutex *sync.Mutex
}

func (canvas *Canvas) FromImage(img *image.Image) {
	canvas.Width = (*img).Bounds().Dx()
	canvas.Height = (*img).Bounds().Dy()
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
	return canvas.Data[(y*canvas.Width+x)*3], canvas.Data[(y*canvas.Width+x)*3+1], canvas.Data[(y*canvas.Width+x)*3+2]
}

func worker(start, stop int) {
	ws, err := connectWs(wsUrl)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		for y := 0; y < place.Height; y++ {
			if y*place.Width < start || y*place.Width >= stop {
				continue
			}

			for x := 0; x < place.Width; x++ {
				time.Sleep(1 * time.Millisecond)

				if y*place.Width+x < start || y*place.Width+x >= stop {
					continue
				}

				cx, cy := offset[0]+x, offset[1]+y

				r, g, b := place.At(x, y)

				canvas.Mutex.Lock()

				cr, cg, cb := canvas.At(cx, cy)

				canvas.Mutex.Unlock()

				if r == cr && g == cg && b == cb {
					continue
				}

				message := make([]byte, 11)

				binary.BigEndian.PutUint32(message[0:4], uint32(cx))
				binary.BigEndian.PutUint32(message[4:8], uint32(cy))

				message[8] = uint8(r)
				message[9] = uint8(g)
				message[10] = uint8(b)

				if err := ws.WriteMessage(websocket.BinaryMessage, message); err != nil {
					fmt.Println(err)
					// try to reconnect

					for {
						if ws, err = connectWs(wsUrl); err == nil {
							fmt.Println("Reconnected to websocket")
							break
						}

						time.Sleep(1 * time.Second)
					}
				}

				// fmt.Println("Placed pixel at", cx, cy)

				// time.Sleep(1500 * time.Microsecond)
			}
		}

	}

}

var canvas = Canvas{Mutex: &sync.Mutex{}}
var masterWs *websocket.Conn
var place = Canvas{}

func main() {
	if len(os.Args) > 3 {
		x, err := strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		y, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Println(err)
			return
		}
		offset = [2]int{x, y}

		imagePath = os.Args[3]
	}

	img, err := getPlacePng(placePngUrl)
	if err != nil {
		fmt.Println(err)
		return
	}
	canvas.FromImage(img)

	fmt.Println("Successfully fetched place.png")

	if masterWs, err = connectWs(wsUrl); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Successfully connected to websocket")

	go func() {
		for {
			// time.Sleep(100 * time.Microsecond)
			_, message, err := masterWs.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}

			x := binary.BigEndian.Uint32(message[0:4])
			y := binary.BigEndian.Uint32(message[4:8])

			r := message[8]
			g := message[9]
			b := message[10]

			canvas.Data[(int(y)*canvas.Width+int(x))*3] = r
			canvas.Data[(int(y)*canvas.Width+int(x))*3+1] = g
			canvas.Data[(int(y)*canvas.Width+int(x))*3+2] = b
		}
	}()

	placeImage, err := loadImage(imagePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	place.FromImage(placeImage)

	fmt.Println("Successfully loaded image to place")

	// time.Sleep(5 * time.Second)

	workers := 10
	pixels := place.Width * place.Height
	perWorker := pixels / workers

	// 10 workers
	for i := 0; i < 9; i++ {
		go worker(i*perWorker, (i+1)*perWorker)
	}

	worker(0, pixels)

}
