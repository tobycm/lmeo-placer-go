package main

import (
	"fmt"
	"strconv"
	"strings"
)

var message = "00 00 01 34 00 00 01 8F 00 00 00"

func main() {
	message = strings.Replace(message, " ", "", -1)

	x, err := strconv.ParseUint(message[0:8], 16, 32)
	if err != nil {
		panic(err)
	}
	y, err := strconv.ParseUint(message[8:16], 16, 32)
	if err != nil {
		panic(err)
	}

	r, err := strconv.ParseUint(message[16:18], 16, 8)
	if err != nil {
		panic(err)
	}
	g, err := strconv.ParseUint(message[18:20], 16, 8)
	if err != nil {
		panic(err)
	}
	b, err := strconv.ParseUint(message[20:22], 16, 8)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Coords: (%d, %d)\n", x, y)
	fmt.Printf("Color: #%02X%02X%02X\n", r, g, b)

}
