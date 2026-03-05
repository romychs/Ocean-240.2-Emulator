package okean240

import "image/color"

var CRed = color.RGBA{R: 0xff, G: 0, B: 0, A: 0xff}
var CLGreen = color.RGBA{R: 0x00, G: 0xff, B: 0, A: 0xff}
var CGreen = color.RGBA{R: 0x12, G: 0x76, B: 0x22, A: 0xff}
var CBlue = color.RGBA{R: 0x2A, G: 0x60, B: 0x99, A: 0xff}
var CLBlue = color.RGBA{R: 0x72, G: 0x9F, B: 0xCF, A: 0xff}
var CCrimson = color.RGBA{R: 0xbf, G: 0x00, B: 0xbf, A: 0xff}
var CWhite = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
var CYellow = color.RGBA{R: 0xff, G: 0xff, B: 0x00, A: 0xff}
var CBlack = color.RGBA{R: 0, G: 0, B: 0, A: 0xff}

//// R5 - Style Palette
//var ColorPalette = [8][4]color.RGBA{
//	{CBlack, CRed, CGreen, CBlue},    // 000
//	{CWhite, CRed, CGreen, CBlue},    // 001
//	{CRed, CGreen, CLBlue, CYellow},  // 010
//	{CBlack, CRed, CCrimson, CWhite}, // 011
//	{CBlack, CRed, CWhite, CBlue},    // 100
//	{CBlack, CRed, CWhite, CYellow},  // 101
//	{CGreen, CWhite, CYellow, CBlue}, // 110
//	{CBlack, CBlack, CBlack, CBlack}, // 111
//}

// ColorPalette R8 style palette
var ColorPalette = [8][4]color.RGBA{
	{CBlack, CGreen, CBlue, CRed},     // 000
	{CWhite, CGreen, CBlue, CRed},     // 001
	{CGreen, CBlue, CCrimson, CLBlue}, // 010
	{CBlack, CGreen, CYellow, CWhite}, // 011
	{CBlack, CGreen, CLBlue, CRed},    // 100
	{CBlack, CRed, CBlue, CLBlue},     // 101
	{CBlue, CWhite, CLBlue, CRed},     // 110
	{CBlack, CBlack, CBlack, CBlack},  // 111
}

// BgColorPalette Background color palette
var BgColorPalette = [8]color.RGBA{
	CBlack, CBlue, CGreen, CLBlue, CRed, CCrimson, CYellow, CWhite,
}

//var MonoPalette = [8]color.RGBA{
//	CWhite, CRed, CGreen, CBlue, CLBlue, CYellow, CLGreen, CBlack,
//}

// MonoPalette R8 Style monochrome mode palette
var MonoPalette = [8]color.RGBA{
	CRed, CBlue, CCrimson, CGreen, CYellow, CLBlue, CWhite, CBlack,
}
