package escpos

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	// ASCII DLE (DataLinkEscape)
	DLE byte = 0x10

	// ASCII EOT (EndOfTransmission)
	EOT byte = 0x04

	// ASCII GS (Group Separator)
	GS byte = 0x1D
)

// text replacement map
var textReplaceMap = map[string]string{
	// horizontal tab
	"&#9;":  "\x09",
	"&#x9;": "\x09",

	// linefeed
	"&#10;": "\n",
	"&#xA;": "\n",

	// xml stuff
	"&apos;": "'",
	"&quot;": `"`,
	"&gt;":   ">",
	"&lt;":   "<",

	// ampersand must be last to avoid double decoding
	"&amp;": "&",
}

// replace text from the above map
func textReplace(data string) string {
	for k, v := range textReplaceMap {
		data = strings.Replace(data, k, v, -1)
	}
	return data
}

func (e *Escpos) SetChineseOn() {
	e.Write(fmt.Sprintf("\x1C&"))
}

type Escpos struct {
	// destination
	dst io.ReadWriter

	// font metrics
	width, height uint8

	// state toggles ESC[char]
	underline  uint8
	emphasize  uint8
	upsidedown uint8
	rotate     uint8

	// state toggles GS[char]
	reverse, smooth uint8
}

// reset toggles
func (e *Escpos) reset() {
	e.width = 1
	e.height = 1

	e.underline = 0
	e.emphasize = 0
	e.upsidedown = 0
	e.rotate = 0

	e.reverse = 0
	e.smooth = 0
}

// create Escpos printer
func New(dst io.ReadWriter) (e *Escpos) {
	e = &Escpos{dst: dst}
	e.reset()
	return
}

// write raw bytes to printer
func (e *Escpos) WriteRaw(data []byte) (n int, err error) {
	if len(data) > 0 {
		e.dst.Write(data)
	}

	return 0, nil
}

// read raw bytes from printer
func (e *Escpos) ReadRaw(data []byte) (n int, err error) {
	return e.dst.Read(data)
}

// write a string to the printer
func (e *Escpos) Write(data string) (int, error) {
	reader := transform.NewReader(bytes.NewReader([]byte(data)), simplifiedchinese.GB18030.NewEncoder())
	bs, _ := ioutil.ReadAll(reader)
	return e.WriteRaw(bs)
}

//开钱箱
func (e *Escpos) OpenDrawer() {
	e.WriteRaw([]byte{0x1b, 0x70, byte(0), byte(10), byte(10)})
}

// init/reset printer settings
func (e *Escpos) Init() {
	e.reset()
	e.Write("\x1B@")
}

// end output
func (e *Escpos) End() {
	e.Write("\xFA")
}

// send cut
func (e *Escpos) Cut() {
	e.Write("\x1DVA0")
}

// send cut minus one point (partial cut)
func (e *Escpos) CutPartial() {
	e.WriteRaw([]byte{GS, 0x56, 1})
}

// send cash
func (e *Escpos) Cash() {
	e.Write("\x1B\x70\x00\x0A\xFF")
}

// send linefeed
func (e *Escpos) Linefeed() {
	e.Write("\n")
}

// send N formfeeds
func (e *Escpos) FormfeedN(n int) {
	e.Write(fmt.Sprintf("\x1Bd%c", n))
}

// send formfeed
func (e *Escpos) Formfeed() {
	e.FormfeedN(1)
}

// set font
func (e *Escpos) SetFont(font string) {
	f := 0

	switch font {
	case "A":
		f = 0
	case "B":
		f = 1
	case "C":
		f = 2
	default:
		f = 0
	}

	e.Write(fmt.Sprintf("\x1BM%c", f))
}

func (e *Escpos) SendFontSize() {
	e.Write(fmt.Sprintf("\x1D!%c", ((e.width)<<4)|(e.height)))
}
func (e *Escpos) SetFontStyle(style uint8) {
	e.Write(string([]byte{0x1b, 0x21, byte(style)}))
}

func (e *Escpos) SetLetterSpace(n int) {
	e.Write(string([]byte{0x1b, 0x20, byte(n)}))
}

// set font size
func (e *Escpos) SetFontSize(width, height uint8) {
	if width >= 0 && height >= 0 && width < 8 && height < 8 {
		e.width = width
		e.height = height
		e.SendFontSize()
	}
}

func (e *Escpos) SetFontColor(color uint8) {
	e.WriteRaw([]byte{0x1b, 0x72, byte(color)})
}

// send underline
func (e *Escpos) SendUnderline() {
	e.Write(fmt.Sprintf("\x1B-%c", e.underline))
}

// send emphasize / doublestrike
func (e *Escpos) SendEmphasize() {
	e.Write(fmt.Sprintf("\x1BG%c", e.emphasize))
}

// send upsidedown
func (e *Escpos) SendUpsidedown() {
	e.Write(fmt.Sprintf("\x1B{%c", e.upsidedown))
}

// send rotate
func (e *Escpos) SendRotate() {
	e.Write(fmt.Sprintf("\x1BR%c", e.rotate))
}

// send reverse
func (e *Escpos) SendReverse() {
	e.Write(fmt.Sprintf("\x1DB%c", e.reverse))
}

// send smooth
func (e *Escpos) SendSmooth() {
	e.Write(fmt.Sprintf("\x1Db%c", e.smooth))
}

// 光标移动到x位置
func (e *Escpos) SendMoveX(x int) {
	e.Write(string([]byte{0x1b, 0x24, byte(x % 256), byte(x / 256)}))
}

// send move y
func (e *Escpos) SendMoveY(y int) {
	e.Write(string([]byte{0x1d, 0x24, byte(y % 256), byte(y / 256)}))
}

// set underline
func (e *Escpos) SetUnderline(v uint8) {
	e.underline = v
	e.SendUnderline()
}

// set emphasize
func (e *Escpos) SetEmphasize(u uint8) {
	e.emphasize = u
	e.SendEmphasize()
}

// set upsidedown
func (e *Escpos) SetUpsidedown(v uint8) {
	e.upsidedown = v
	e.SendUpsidedown()
}

// set rotate
func (e *Escpos) SetRotate(v uint8) {
	e.rotate = v
	e.SendRotate()
}

// set reverse
func (e *Escpos) SetReverse(v uint8) {
	e.reverse = v
	e.SendReverse()
}

// set smooth
func (e *Escpos) SetSmooth(v uint8) {
	e.smooth = v
	e.SendSmooth()
}

// pulse (open the drawer)
func (e *Escpos) Pulse() {
	// with t=2 -- meaning 2*2msec
	e.Write("\x1Bp\x02")
}

// set alignment
func (e *Escpos) SetAlign(align string) {
	a := 0
	switch align {
	case "left":
		a = 0
	case "center":
		a = 1
	case "right":
		a = 2
	}
	e.Write(fmt.Sprintf("\x1Ba%c", a))
}

func (e *Escpos) SetMarginLeft(size uint16) {
	if size <= 47 {
		e.Write(string([]byte{0x1d, 0x4c, byte(size % 256), byte(size / 256)}))
	}
}

// set language -- ESC R
func (e *Escpos) SetLang(lang string) {
	l := 0

	switch lang {
	case "en":
		l = 0
	case "fr":
		l = 1
	case "de":
		l = 2
	case "uk":
		l = 3
	case "da":
		l = 4
	case "sv":
		l = 5
	case "it":
		l = 6
	case "es":
		l = 7
	case "ja":
		l = 8
	case "no":
		l = 9
	}
	e.Write(fmt.Sprintf("\x1BR%c", l))
}

// feed and cut based on parameters
func (e *Escpos) FeedAndCut(params map[string]string) {
	if t, ok := params["type"]; ok && t == "feed" {
		e.Formfeed()
	}

	e.Cut()
}

// Barcode sends a barcode to the printer.
func (e *Escpos) Barcode(barcode string, format int) {
	code := ""
	switch format {
	case 0:
		code = "\x00"
	case 1:
		code = "\x01"
	case 2:
		code = "\x02"
	case 3:
		code = "\x03"
	case 4:
		code = "\x04"
	case 73:
		code = "\x49"
	}

	// reset settings
	e.reset()

	// set align
	e.SetAlign("center")

	// write barcode
	if format > 69 {
		e.Write(fmt.Sprintf("\x1dk"+code+"%v%v", len(barcode), barcode))
	} else if format < 69 {
		e.Write(fmt.Sprintf("\x1dk"+code+"%v\x00", barcode))
	}
	e.Write(fmt.Sprintf("%v", barcode))
}

// used to send graphics headers
func (e *Escpos) gSend(m byte, fn byte, data []byte) {
	l := len(data) + 2

	e.Write("\x1b(L")
	e.WriteRaw([]byte{byte(l % 256), byte(l / 256), m, fn})
	e.WriteRaw(data)
}

// ReadStatus Read the status n from the printer
func (e *Escpos) ReadStatus(n byte) (byte, error) {
	e.WriteRaw([]byte{DLE, EOT, n})
	data := make([]byte, 1)
	_, err := e.ReadRaw(data)
	if err != nil {
		return 0, err
	}
	return data[0], nil
}
