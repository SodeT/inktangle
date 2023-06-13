package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

var pl = fmt.Println
var pf = fmt.Printf

type pixel_pair struct {
	px1		color.NRGBA
	px2		color.NRGBA
}

func main() {

	argc := len(os.Args)
	if argc < 3 {
		panic("Too few arguments...")
	}

	switch arg1 := os.Args[1]; arg1 {
	case "w":
		var msg string
		for i := 3; i < argc; i++ {
			msg += os.Args[i] + " "
		}
		msg += "\\"
		encode_message(msg, os.Args[2])
	case "r":
		msg := string(decode_message(os.Args[2]))
		pl(get_readable(msg))
	default:
		pl("Wrong input...")
	}

}

func get_readable(msg string) (readable string) {
	for i := 0; i < len(msg); i++ {
		if msg[i] == '\\' {
			return readable
		}
		readable += string(msg[i])
	}
	return "Error..."
}

func encode_message(msg string, file string) {
	src_img, err := read_image(file)
	if err != nil {
		log.Fatal(err)
	}

	width := src_img.Bounds().Max.X
	height := src_img.Bounds().Max.Y

    dst_img := image.NewNRGBA(image.Rect(0, 0, width, height))

	msg += "\\"

	for i := 0; i < width * height -1; i++ {
		

		x1, y1 := to_pos(i*2, width)
		x2, y2 := to_pos(i*2+1, width)
	
		var pair pixel_pair
		pair.px1 = src_img.NRGBAAt(x1, y1)
		pair.px2 = src_img.NRGBAAt(x2, y2)

		if i >= len(msg) {
			dst_img.SetNRGBA(x1, y1, pair.px1)
			dst_img.SetNRGBA(x2, y2, pair.px2)
			//TODO pixel lowest right does not get drawn if the pixels count is uneven
			continue
		}


		new_pair := encode_char(pair, msg[i])
		dst_img.SetNRGBA(x1, y1, new_pair.px1)
		dst_img.SetNRGBA(x2, y2, new_pair.px2)
	}

	write_img(*dst_img, "INKT_" + file)
}

func encode_char(pair pixel_pair, char byte) (new_pair pixel_pair) {
	
	r := pair.px1.R
	g := pair.px1.G
	b := pair.px1.B
	a := pair.px1.A

	r, char = write_bit(r, char)
	g, char = write_bit(g, char)
	b, char = write_bit(b, char)
	a, char = write_bit(a, char)
	
	new_pair.px1 = color.NRGBA{r, g, b, a}
	
	r2 := pair.px1.R
	g2 := pair.px1.G
	b2 := pair.px1.B
	a2 := pair.px1.A

	r2, char = write_bit(r2, char)
	g2, char = write_bit(g2, char)
	b2, char = write_bit(b2, char)
	a2, char = write_bit(a2, char)
	
	new_pair.px2 = color.NRGBA{r2, g2, b2, a2}
	
	return new_pair
}

func write_bit(channel uint8, char byte) (uint8, byte) {
	last_char_bit := char & 1		// get the bit we want to write to the channel
	last_channel_bit := channel & 1	// get the bit that is currently set in the channel 
	if last_channel_bit != last_char_bit {
		channel = channel ^ 1		// change the bit if the current bit in the channel is wrong
	}
	char = char >> 1				// shift the byte one step to acces the next bit
	return channel, char
}

func decode_message(file string) []byte {
	img, err := read_image(file)
	if err != nil {
		log.Fatal(err)
	}

	width := img.Bounds().Max.X
	height := img.Bounds().Max.Y

	var msg []byte
	for i := 0; i < width * height -1; i += 2 {
		x1, y1 := to_pos(i, width)
		x2, y2 := to_pos(i+1, width)

		var pair pixel_pair
		pair.px1 = img.NRGBAAt(x1, y1)
		pair.px2 = img.NRGBAAt(x2, y2)


		msg = append(msg, decode_char(pair))
	}

	return msg
}


func decode_char(pair pixel_pair) (char byte) {
	r := pair.px1.R
	g := pair.px1.G
	b := pair.px1.B
	a := pair.px1.A

	r2 := pair.px2.R
	g2 := pair.px2.G
	b2 := pair.px2.B
	a2 := pair.px2.A

	// read order is important and is the write order in reverse
	char = read_bit(a2, char)
	char = read_bit(b2, char)
	char = read_bit(g2, char)
	char = read_bit(r2, char)
	
	char = read_bit(a, char)
	char = read_bit(b, char)
	char = read_bit(g, char)
	char = read_bit(r, char)
	return char
}


func read_bit(channel uint8, char byte) byte {
	char = char << 1				// shift char to write next bit to output
	last_channel_bit := channel & 1	// get the last bit of the channel 
	char = char ^ last_channel_bit	// set the char bit
	return char
}

func to_pos(i int, width int) (x int, y int) {
	x = i % width
	y = i / width
	return x, y
}

func read_image(path string) (image.NRGBA, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	
	img_data, err := png.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	if img_nrgba, ok := img_data.(*image.NRGBA); ok {
		return *img_nrgba, nil
	}
	return *image.NewNRGBA(image.Rect(0,0,0,0)), errors.New("File could not be converted...")
}

func write_img(data image.NRGBA, path string) {
	out, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	png.Encode(out, data.SubImage(image.Rect(0,0,data.Bounds().Max.X,data.Bounds().Max.Y)))
	out.Close()
	return
}
