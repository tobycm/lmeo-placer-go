package main

import (
	"encoding/binary"
	"fmt"
	"image"
	_ "image/png"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func GetCanvasImage(url string) (*image.Image, error) {
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

type PlaceWs struct {
	Conn *websocket.Conn

	Url           string
	AutoReconnect bool
}

func (p *PlaceWs) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(p.Url, nil)
	if err != nil {
		if p.AutoReconnect {
			p.Reconnect()
			return nil
		}

		return fmt.Errorf("error connecting to websocket: %w", err)
	}

	p.Conn = conn

	return nil
}

func (p *PlaceWs) Reconnect() {
	for {
		if err := p.Connect(); err == nil {
			fmt.Println("Reconnected to websocket")
			return
		}

		time.Sleep(1 * time.Second)
	}
}

func (p *PlaceWs) PlacePixel(x, y int, r, g, b byte) error {
	message := make([]byte, 11)

	binary.BigEndian.PutUint32(message[0:4], uint32(x))
	binary.BigEndian.PutUint32(message[4:8], uint32(y))

	message[8] = r
	message[9] = g
	message[10] = b

	if err := p.Conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
		if p.AutoReconnect {
			p.Reconnect()
		}

		return fmt.Errorf("error writing message to websocket: %w", err)
	}

	return nil
}
