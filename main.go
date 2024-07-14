package main

import (
	"encoding/binary"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var placePngUrl = "https://foloplace.tobycm.dev/place.png"
var wsUrl = "wss://foloplace.tobycm.dev/ws"
var imagePath = "./elysia smol.png"

var offset = [2]int{200, 306} // starting point, [x, y]

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

var canvas = Canvas{}
var ws *websocket.Conn
var place = Canvas{}

func main() {
	img, err := getPlacePng(placePngUrl)
	if err != nil {
		fmt.Println(err)
		return
	}
	canvas.FromImage(img)

	fmt.Println("Successfully fetched place.png")

	if ws, err = connectWs(wsUrl); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Successfully connected to websocket")
	defer ws.Close()

	go func() {
		for {
			_, message, err := ws.ReadMessage()
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

	for {
		for y := 0; y < place.Height; y++ {
			for x := 0; x < place.Width; x++ {

				cx, cy := offset[0]+x, offset[1]+y

				r, g, b := place.At(x, y)
				cr, cg, cb := canvas.At(cx, cy)

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
					ws.Close()

					for {
						if ws, err = connectWs(wsUrl); err != nil {
							fmt.Println(err)
							time.Sleep(2 * time.Second)
							continue
						}
						break
					}

					return
				}

				time.Sleep(1500 * time.Microsecond)
			}
		}

	}
}
