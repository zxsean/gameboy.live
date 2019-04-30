package driver

import (
	"log"
	"math"
	"net"
)

type ASCII struct {
	// Origin pixel data generated by the emulator
	pixels *[160][144][3]uint8
	Conn   net.Conn
	// How many frames have been sent by the emulator
	FrameCount int
	// Last sent data ,used for comparing with the next frame
	last  [160][144]bool
	title string
}

func (stream *ASCII) Init(pixels *[160][144][3]uint8, title string) {
	stream.title = title
	stream.pixels = pixels

}

func (stream *ASCII) Run(drawSignal chan bool) {

	for {
		if !<-drawSignal {
			log.Println("chan closed")
			break
		}
		stream.FrameCount++
		pixels := [160][144]bool{}
		for y := 0; y < 144; y++ {
			for x := 0; x < 160; x++ {
				switch stream.pixels[x][y][0] {
				case 255, 0xCC:
					// White pixel
					pixels[x][y] = true
				default:
					// Black pixel
					pixels[x][y] = false
				}
			}
		}
		stream.renderAscii(pixels)

	}
}

/*
	Render pixels as Braille
	Reference: https://github.com/gabrielrcouto/php-terminal-gameboy-emulator/blob/master/src/Canvas/TerminalCanvas.php
*/
func (stream *ASCII) renderAscii(pixels [160][144]bool) {
	if stream.last == pixels {
		return
	}
	stream.last = pixels
	pixelMap := [][2]uint16{
		{0x2801, 0x2808},
		{0x2802, 0x2810},
		{0x2804, 0x2820},
		{0x2840, 0x2880},
	}
	chars := [2880]uint16{}
	for i := 0; i < 2880; i++ {
		chars[i] = 0x2880
	}
	chars[0] |= pixelMap[0][0]
	chars[2879] |= pixelMap[0][0]
	ret := ""
	for y := 0; y < 144; y++ {
		for x := 0; x < 160; x++ {
			charPosition := int(math.Floor(float64(x)/2.0) + (math.Floor(float64(y)/4.0) * 80))
			if pixels[x][y] {
				chars[charPosition] |= pixelMap[y%4][x%2]
			}
			if x%2 == 1 && y%4 == 3 {
				if chars[charPosition] == 0x2880 {
					ret += " "
				} else {
					ret += string(chars[charPosition])
				}
				if x%159 == 0 {
					ret += "\r\n"
				}
			}
		}
	}
	// Clean screen
	_, err := stream.Conn.Write([]byte("\033[H" + ret))

	if err != nil {
		log.Println("Failed to send frame to player")
	}

}
