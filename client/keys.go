package client

import (
	"fmt"
)

type Key int

const (
	KeyA              Key = 0
	KeyB                  = 1
	KeyC                  = 2
	KeyD                  = 3
	KeyE                  = 4
	KeyF                  = 5
	KeyG                  = 6
	KeyH                  = 7
	KeyI                  = 8
	KeyJ                  = 9
	KeyK                  = 10
	KeyL                  = 11
	KeyM                  = 12
	KeyN                  = 13
	KeyO                  = 14
	KeyP                  = 15
	KeyQ                  = 16
	KeyR                  = 17
	KeyS                  = 18
	KeyT                  = 19
	KeyU                  = 20
	KeyV                  = 21
	KeyW                  = 22
	KeyX                  = 23
	KeyY                  = 24
	KeyZ                  = 25
	KeyAltLeft            = 26
	KeyAltRight           = 27
	KeyArrowDown          = 28
	KeyArrowLeft          = 29
	KeyArrowRight         = 30
	KeyArrowUp            = 31
	KeyBackquote          = 32
	KeyBackslash          = 33
	KeyBackspace          = 34
	KeyBracketLeft        = 35
	KeyBracketRight       = 36
	KeyCapsLock           = 37
	KeyComma              = 38
	KeyContextMenu        = 39
	KeyControlLeft        = 40
	KeyControlRight       = 41
	KeyDelete             = 42
	KeyDigit0             = 43
	KeyDigit1             = 44
	KeyDigit2             = 45
	KeyDigit3             = 46
	KeyDigit4             = 47
	KeyDigit5             = 48
	KeyDigit6             = 49
	KeyDigit7             = 50
	KeyDigit8             = 51
	KeyDigit9             = 52
	KeyEnd                = 53
	KeyEnter              = 54
	KeyEqual              = 55
	KeyEscape             = 56
	KeyF1                 = 57
	KeyF2                 = 58
	KeyF3                 = 59
	KeyF4                 = 60
	KeyF5                 = 61
	KeyF6                 = 62
	KeyF7                 = 63
	KeyF8                 = 64
	KeyF9                 = 65
	KeyF10                = 66
	KeyF11                = 67
	KeyF12                = 68
	KeyF13                = 69
	KeyF14                = 70
	KeyF15                = 71
	KeyF16                = 72
	KeyF17                = 73
	KeyF18                = 74
	KeyF19                = 75
	KeyF20                = 76
	KeyF21                = 77
	KeyF22                = 78
	KeyF23                = 79
	KeyF24                = 80
	KeyHome               = 81
	KeyInsert             = 82
	KeyIntlBackslash      = 83
	KeyMetaLeft           = 84
	KeyMetaRight          = 85
	KeyMinus              = 86
	KeyNumLock            = 87
	KeyNumpad0            = 88
	KeyNumpad1            = 89
	KeyNumpad2            = 90
	KeyNumpad3            = 91
	KeyNumpad4            = 92
	KeyNumpad5            = 93
	KeyNumpad6            = 94
	KeyNumpad7            = 95
	KeyNumpad8            = 96
	KeyNumpad9            = 97
	KeyNumpadAdd          = 98
	KeyNumpadDecimal      = 99
	KeyNumpadDivide       = 100
	KeyNumpadEnter        = 101
	KeyNumpadEqual        = 102
	KeyNumpadMultiply     = 103
	KeyNumpadSubtract     = 104
	KeyPageDown           = 105
	KeyPageUp             = 106
	KeyPause              = 107
	KeyPeriod             = 108
	KeyPrintScreen        = 109
	KeyQuote              = 110
	KeyScrollLock         = 111
	KeySemicolon          = 112
	KeyShiftLeft          = 113
	KeyShiftRight         = 114
	KeySlash              = 115
	KeySpace              = 116
	KeyTab                = 117
	KeyReserved0          = 118
	KeyReserved1          = 119
	KeyReserved2          = 120
	KeyReserved3          = 121
	KeyMax                = KeyReserved3
)

func (k Key) String() string {
	switch k {
	case KeyA:
		return "KeyA"
	case KeyB:
		return "KeyB"
	case KeyC:
		return "KeyC"
	case KeyD:
		return "KeyD"
	case KeyE:
		return "KeyE"
	case KeyF:
		return "KeyF"
	case KeyG:
		return "KeyG"
	case KeyH:
		return "KeyH"
	case KeyI:
		return "KeyI"
	case KeyJ:
		return "KeyJ"
	case KeyK:
		return "KeyK"
	case KeyL:
		return "KeyL"
	case KeyM:
		return "KeyM"
	case KeyN:
		return "KeyN"
	case KeyO:
		return "KeyO"
	case KeyP:
		return "KeyP"
	case KeyQ:
		return "KeyQ"
	case KeyR:
		return "KeyR"
	case KeyS:
		return "KeyS"
	case KeyT:
		return "KeyT"
	case KeyU:
		return "KeyU"
	case KeyV:
		return "KeyV"
	case KeyW:
		return "KeyW"
	case KeyX:
		return "KeyX"
	case KeyY:
		return "KeyY"
	case KeyZ:
		return "KeyZ"
	case KeyAltLeft:
		return "KeyAltLeft"
	case KeyAltRight:
		return "KeyAltRight"
	case KeyArrowDown:
		return "KeyArrowDown"
	case KeyArrowLeft:
		return "KeyArrowLeft"
	case KeyArrowRight:
		return "KeyArrowRight"
	case KeyArrowUp:
		return "KeyArrowUp"
	case KeyBackquote:
		return "KeyBackquote"
	case KeyBackslash:
		return "KeyBackslash"
	case KeyBackspace:
		return "KeyBackspace"
	case KeyBracketLeft:
		return "KeyBracketLeft"
	case KeyBracketRight:
		return "KeyBracketRight"
	case KeyCapsLock:
		return "KeyCapsLock"
	case KeyComma:
		return "KeyComma"
	case KeyContextMenu:
		return "KeyContextMenu"
	case KeyControlLeft:
		return "KeyControlLeft"
	case KeyControlRight:
		return "KeyControlRight"
	case KeyDelete:
		return "KeyDelete"
	case KeyDigit0:
		return "KeyDigit0"
	case KeyDigit1:
		return "KeyDigit1"
	case KeyDigit2:
		return "KeyDigit2"
	case KeyDigit3:
		return "KeyDigit3"
	case KeyDigit4:
		return "KeyDigit4"
	case KeyDigit5:
		return "KeyDigit5"
	case KeyDigit6:
		return "KeyDigit6"
	case KeyDigit7:
		return "KeyDigit7"
	case KeyDigit8:
		return "KeyDigit8"
	case KeyDigit9:
		return "KeyDigit9"
	case KeyEnd:
		return "KeyEnd"
	case KeyEnter:
		return "KeyEnter"
	case KeyEqual:
		return "KeyEqual"
	case KeyEscape:
		return "KeyEscape"
	case KeyF1:
		return "KeyF1"
	case KeyF2:
		return "KeyF2"
	case KeyF3:
		return "KeyF3"
	case KeyF4:
		return "KeyF4"
	case KeyF5:
		return "KeyF5"
	case KeyF6:
		return "KeyF6"
	case KeyF7:
		return "KeyF7"
	case KeyF8:
		return "KeyF8"
	case KeyF9:
		return "KeyF9"
	case KeyF10:
		return "KeyF10"
	case KeyF11:
		return "KeyF11"
	case KeyF12:
		return "KeyF12"
	case KeyF13:
		return "KeyF13"
	case KeyF14:
		return "KeyF14"
	case KeyF15:
		return "KeyF15"
	case KeyF16:
		return "KeyF16"
	case KeyF17:
		return "KeyF17"
	case KeyF18:
		return "KeyF18"
	case KeyF19:
		return "KeyF19"
	case KeyF20:
		return "KeyF20"
	case KeyF21:
		return "KeyF21"
	case KeyF22:
		return "KeyF22"
	case KeyF23:
		return "KeyF23"
	case KeyF24:
		return "KeyF24"
	case KeyHome:
		return "KeyHome"
	case KeyInsert:
		return "KeyInsert"
	case KeyIntlBackslash:
		return "KeyIntlBackslash"
	case KeyMetaLeft:
		return "KeyMetaLeft"
	case KeyMetaRight:
		return "KeyMetaRight"
	case KeyMinus:
		return "KeyMinus"
	case KeyNumLock:
		return "KeyNumLock"
	case KeyNumpad0:
		return "KeyNumpad0"
	case KeyNumpad1:
		return "KeyNumpad1"
	case KeyNumpad2:
		return "KeyNumpad2"
	case KeyNumpad3:
		return "KeyNumpad3"
	case KeyNumpad4:
		return "KeyNumpad4"
	case KeyNumpad5:
		return "KeyNumpad5"
	case KeyNumpad6:
		return "KeyNumpad6"
	case KeyNumpad7:
		return "KeyNumpad7"
	case KeyNumpad8:
		return "KeyNumpad8"
	case KeyNumpad9:
		return "KeyNumpad9"
	case KeyNumpadAdd:
		return "KeyNumpadAdd"
	case KeyNumpadDecimal:
		return "KeyNumpadDecimal"
	case KeyNumpadDivide:
		return "KeyNumpadDivide"
	case KeyNumpadEnter:
		return "KeyNumpadEnter"
	case KeyNumpadEqual:
		return "KeyNumpadEqual"
	case KeyNumpadMultiply:
		return "KeyNumpadMultiply"
	case KeyNumpadSubtract:
		return "KeyNumpadSubtract"
	case KeyPageDown:
		return "KeyPageDown"
	case KeyPageUp:
		return "KeyPageUp"
	case KeyPause:
		return "KeyPause"
	case KeyPeriod:
		return "KeyPeriod"
	case KeyPrintScreen:
		return "KeyPrintScreen"
	case KeyQuote:
		return "KeyQuote"
	case KeyScrollLock:
		return "KeyScrollLock"
	case KeySemicolon:
		return "KeySemicolon"
	case KeyShiftLeft:
		return "KeyShiftLeft"
	case KeyShiftRight:
		return "KeyShiftRight"
	case KeySlash:
		return "KeySlash"
	case KeySpace:
		return "KeySpace"
	case KeyTab:
		return "KeyTab"
	}

	return fmt.Sprintf("Key(%d)", k)
}
