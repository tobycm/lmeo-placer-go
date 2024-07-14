package main

import (
	"strconv"
)

func parseArgs(args []string) ([2]int, string, error) {
	x, err := strconv.Atoi(args[1])
	if err != nil {
		return [2]int{0, 0}, "", err
	}
	y, err := strconv.Atoi(args[2])
	if err != nil {
		return [2]int{0, 0}, "", err
	}

	offset := [2]int{x, y}

	imagePath := args[3]

	return offset, imagePath, nil
}
