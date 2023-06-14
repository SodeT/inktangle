package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

var pl = fmt.Println
var pf = fmt.Printf

var esc_char byte = '\\'

type pixel_pair struct {
	px1		color.NRGBA
	px2		color.NRGBA
}

func main() {
	var usage string = "Usage: inktangle [w | r] [path] [message]"

	argc := len(os.Args)
	if argc < 3 {
		pl(usage)
		return
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
		pl(msg)
	default:
		pl(usage)
	}
	return
}

func encode_message(msg string, file string) {

	// read src file 
	img := read_image(file)

	width := img.Bounds().Max.X
	height := img.Bounds().Max.Y
	pixel_count := width * height

	msg += string(esc_char)

	if len(msg) >= pixel_count / 2 {
		log.Fatal("Message does not fit in this image...")
	}

	interval := pixel_count / len(msg)
	interval %= 255
	
	x1, y1 := to_pos(0, width)
	x2, y2 := to_pos(1, width)

	var pair pixel_pair
	pair.px1 = img.NRGBAAt(x1, y1)
	pair.px2 = img.NRGBAAt(x2, y2)

	new_pair := encode_char(pair, byte(interval))
	img.SetNRGBA(x1, y1, new_pair.px1)
	img.SetNRGBA(x2, y2, new_pair.px2)

	msg_index := 0
	
	// write and encode msg
	for i := interval; i < pixel_count -1; i += interval {
	
		x1, y1 := to_pos(i, width)
		x2, y2 := to_pos(i+1, width)
		
		if msg_index >= len(msg) {
			break
		}
	
		var pair pixel_pair
		pair.px1 = img.NRGBAAt(x1, y1)
		pair.px2 = img.NRGBAAt(x2, y2)


		new_pair := encode_char(pair, msg[msg_index])
		img.SetNRGBA(x1, y1, new_pair.px1)
		img.SetNRGBA(x2, y2, new_pair.px2)
		msg_index++
	}

	// save output
	write_img(img, "INK_" + file)
}

func encode_char(pair pixel_pair, char byte) (new_pair pixel_pair) {
	
	r := pair.px1.R
	g := pair.px1.G
	b := pair.px1.B
	a := pair.px1.A

	r2 := pair.px1.R
	g2 := pair.px1.G
	b2 := pair.px1.B
	a2 := pair.px1.A
	
	// the order the bits are written in matter
	r, char = write_bit(r, char)
	g, char = write_bit(g, char)
	b, char = write_bit(b, char)
	a, char = write_bit(a, char)
	
	r2, char = write_bit(r2, char)
	g2, char = write_bit(g2, char)
	b2, char = write_bit(b2, char)
	a2, char = write_bit(a2, char)
	
	new_pair.px1 = color.NRGBA{r, g, b, a}
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
	// read src file
	img := read_image(file)

	width := img.Bounds().Max.X
	height := img.Bounds().Max.Y

	
	x1, y1 := to_pos(0, width)
	x2, y2 := to_pos(1, width)

	var pair pixel_pair
	pair.px1 = img.NRGBAAt(x1, y1)
	pair.px2 = img.NRGBAAt(x2, y2)

	var interval int = int(decode_char(pair))

	// loop through image pixels
	var msg []byte
	for i := interval; i < width * height -1; i += interval {
		x1, y1 := to_pos(i, width)
		x2, y2 := to_pos(i+1, width)

		var pair pixel_pair
		pair.px1 = img.NRGBAAt(x1, y1)
		pair.px2 = img.NRGBAAt(x2, y2)

		var char byte = decode_char(pair)
		if char == esc_char {
			return msg
		}
		msg = append(msg, char)
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

func read_image(path string) image.NRGBA {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	
	img_data, err := png.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	img_nrgba, ok := img_data.(*image.NRGBA)
	if !ok {
		log.Fatal("File could not be converted...")
	}
	return *img_nrgba
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
