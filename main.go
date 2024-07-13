package main

import (
	"encoding/binary"
	"fmt"
	"image"
	_ "image/png"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

var placePngUrl = "https://foloplace.tobycm.dev/place.png"
var wsUrl = "wss://foloplace.tobycm.dev/ws"
var imagePath = "./dacloud 256.png"

var offset = [2]int{200, 50} // starting point, [x, y]

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
	Data   *image.Image
}

var canvas Canvas
var ws *websocket.Conn
var place Canvas

func main() {
	img, err := getPlacePng(placePngUrl)
	if err != nil {
		fmt.Println(err)
		return
	}

	canvas = Canvas{
		Width:  (*img).Bounds().Dx(),
		Height: (*img).Bounds().Dy(),
		Data:   img,
	}

	fmt.Println("Successfully fetched place.png")

	if ws, err = connectWs(wsUrl); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Successfully connected to websocket")
	defer ws.Close()

	placeImage, err := loadImage(imagePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	place = Canvas{
		Width:  (*placeImage).Bounds().Dx(),
		Height: (*placeImage).Bounds().Dy(),
		Data:   placeImage,
	}

	fmt.Println("Successfully loaded image to place")

	// time.Sleep(5 * time.Second)

	for y := 0; y < place.Height; y++ {
		for x := 0; x < place.Width; x++ {

			r, g, b, _ := (*place.Data).At(x, y).RGBA()
			r = r >> 8
			g = g >> 8
			b = b >> 8

			cr, cg, cb, _ := (*canvas.Data).At(offset[0]+x, offset[1]+y).RGBA()
			cr = cr >> 8
			cg = cg >> 8
			cb = cb >> 8

			if r == cr && g == cg && b == cb {
				continue
			}

			message := make([]byte, 11)

			binary.BigEndian.PutUint32(message[0:4], uint32(offset[0]+x))
			binary.BigEndian.PutUint32(message[4:8], uint32(offset[1]+y))

			message[8] = uint8(r)
			message[9] = uint8(g)
			message[10] = uint8(b)

			if err := ws.WriteMessage(websocket.BinaryMessage, message); err != nil {
				fmt.Println(err)
				return
			}

			time.Sleep(1 * time.Millisecond)
		}
	}
}
