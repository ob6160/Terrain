package utils

import (
	"bufio"
	"log"
	"math/rand"
	"os"
)

func ReadTextFile(path string) (body string, err error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		body += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return body, err
}


func ToIndex(x, y, width int) int {
	return x * width + y
}


// TODO: Refactor out.
type Point struct {
	X, Y int
}

// Deprecated
// TODO: Refactor out.
func (p Point) ToIndex(width int) int {
	return ToIndex(p.X, p.Y, width)
}

type Rectangle struct {
	TopLeft, TopRight, BottomLeft, BottomRight int
}

func Midpoint(p1, p2 int) int {
	return (p2 + p1) / 2
}

func Average(nums ...float32) float32 {
	var total float32 = 0.0
	var count float32 = 0.0
	for _, num := range nums {
		total += num
		count++
	}
	return total / count
}

func Jitter(value, scale float32) float32 {
	random := rand.Float32() * scale * 2
	shift := scale - random
	return shift + value
}