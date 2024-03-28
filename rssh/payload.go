package rssh

import "fmt"

func ParseNextString(str []byte, startIndex int) (string, int, error) {
	strLen, stopIndex := ParseNextUint32(str, startIndex)
	strBytes := str[stopIndex:][:][:strLen]
	stopIndex = stopIndex + int(strLen)

	return string(strBytes), stopIndex, nil
}

func ParseNextUint32(str []byte, startIndex int) (uint32, int) {
	value := ArrayToUint32([4]byte(str[startIndex : startIndex+4]))
	stopIndex := startIndex + 4

	return value, stopIndex
}

type PtyRequest struct {
	Term    string
	Columns uint32
	Rows    uint32
	Width   uint32
	Height  uint32
	Modes   TerminalModes
}

func ParsePtyRequest(req []byte) (*PtyRequest, error) {
	termstr, stopIndex, err := ParseNextString(req, 0)
	if err != nil {
		return nil, err
	}

	columns, stopIndex := ParseNextUint32(req, stopIndex)
	rows, stopIndex := ParseNextUint32(req, stopIndex)
	width, stopIndex := ParseNextUint32(req, stopIndex)
	height, stopIndex := ParseNextUint32(req, stopIndex)
	modes := ParseTerminalModes(req[stopIndex:])

	ptyReq := &PtyRequest{
		Term:    termstr,
		Columns: columns,
		Rows:    rows,
		Width:   width,
		Height:  height,
		Modes:   modes,
	}

	return ptyReq, nil
}

type DirectTCPIP struct {
	HostToConnect     string
	PortToConnect     uint32
	OriginatorAddress string
	OriginatorPort    uint32
}

func ParseDirectTCPIP(req []byte) (*DirectTCPIP, error) {
	hostToConnect, stopIndex, err := ParseNextString(req, 0)
	if err != nil {
		return nil, err
	}
	portToConnect, stopIndex := ParseNextUint32(req, stopIndex)
	originatorAddress, stopIndex, err := ParseNextString(req, stopIndex)
	if err != nil {
		return nil, err
	}
	originatorPort, _ := ParseNextUint32(req, stopIndex)

	directTCPIP := &DirectTCPIP{
		HostToConnect:     hostToConnect,
		PortToConnect:     portToConnect,
		OriginatorAddress: originatorAddress,
		OriginatorPort:    originatorPort,
	}

	return directTCPIP, nil
}

type TerminalModes map[TerminalMode]uint32

func (tm TerminalModes) String() string {
	str := ""
	for k, v := range tm {
		str += fmt.Sprintf("%s: %d\n", k, v)
	}
	return str
}

func (tm TerminalModes) Bytes() []byte {
	modesBytes := []byte{0, 0, 0, 0}

	modesLen := 1
	for k, v := range tm {
		modesBytes = append(modesBytes, byte(k))
		value := Uint32ToArray(v)
		modesBytes = append(modesBytes, value[:]...)
		modesLen += 5
	}

	modesBytes = append(modesBytes, byte(TtyOpEnd))
	modesLenBytes := Uint32ToArray(uint32(modesLen))
	modesBytes[0] = modesLenBytes[0]
	modesBytes[1] = modesLenBytes[1]
	modesBytes[2] = modesLenBytes[2]
	modesBytes[3] = modesLenBytes[3]

	return modesBytes
}

func ParseTerminalModes(modes []byte) TerminalModes {
	tm := make(TerminalModes)

	modesBytes := modes[4:]

	i := 0
	op := modesBytes[i]

	for op != byte(TtyOpEnd) {
		tm[TerminalMode(op)] = ArrayToUint32([4]byte(modesBytes[i+1 : i+5]))

		i += 5
		op = modesBytes[i]
	}

	return tm
}

func ArrayToUint32(b [4]byte) uint32 {
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func Uint32ToArray(i uint32) [4]byte {
	return [4]byte{byte(i >> 24), byte(i >> 16), byte(i >> 8), byte(i)}
}

type TerminalMode byte

const (
	TtyOpEnd TerminalMode = iota
	VIntr
	VQuit
	VErase
	VKill
	VEof
	VEol
	VEol2
	VStart
	VStop
	VSusp
	VDsusp
	VReprint
	VWErase
	VLNExt
	VFLush
	VSwtch
	VStatus
	VDiscard
)

const (
	IgnPar = iota + 30
	ParMrk
	InpCk
	IStrip
	InlCr
	IgnCr
	ICrnl
	IUclc
	IXon
	IXany
	IXoff
	IMaxbel
	IUTF8
)

const (
	ISig = iota + 50
	ICanon
	XCase
	Echo
	EchoE
	EchoK
	EchoNL
	NoFlsh
	ToStop
	IExten
	EChoCtl
	EchoKe
	PendIn
)

const (
	Opost = iota + 70
	OlCuc
	ONlCr
	OcrNl
	OnoCr
	OnlRet
)

const (
	Cs7 = iota + 90
	Cs8
	ParEnb
	ParOdd
)

const (
	TtyOpInputSpeed  = 128
	TtyOpOutputSpeed = 129
)

func (tm TerminalMode) String() string {
	switch tm {
	case TtyOpEnd:
		return "TtyOpEnd"
	case VIntr:
		return "VIntr"
	case VQuit:
		return "VQuit"
	case VErase:
		return "VErase"
	case VKill:
		return "VKill"
	case VEof:
		return "VEof"
	case VEol:
		return "VEol"
	case VEol2:
		return "VEol2"
	case VStart:
		return "VStart"
	case VStop:
		return "VStop"
	case VSusp:
		return "VSusp"
	case VDsusp:
		return "VDsusp"
	case VReprint:
		return "VReprint"
	case VWErase:
		return "VWErase"
	case VLNExt:
		return "VLNExt"
	case VFLush:
		return "VFLush"
	case VSwtch:
		return "VSwtch"
	case VStatus:
		return "VStatus"
	case VDiscard:
		return "VDiscard"
	case IgnPar:
		return "IgnPar"
	case ParMrk:
		return "ParMrk"
	case InpCk:
		return "InpCk"
	case IStrip:
		return "IStrip"
	case InlCr:
		return "InlCr"
	case IgnCr:
		return "IgnCr"
	case ICrnl:
		return "ICrnl"
	case IUclc:
		return "IUclc"
	case IXon:
		return "IXon"
	case IXany:
		return "IXany"
	case IXoff:
		return "IXoff"
	case IMaxbel:
		return "IMaxbel"
	case IUTF8:
		return "IUTF8"
	case ISig:
		return "ISig"
	case ICanon:
		return "ICanon"
	case XCase:
		return "XCase"
	case Echo:
		return "Echo"
	case EchoE:
		return "EchoE"
	case EchoK:
		return "EchoK"
	case EchoNL:
		return "EchoNL"
	case NoFlsh:
		return "NoFlsh"
	case ToStop:
		return "ToStop"
	case IExten:
		return "IExten"
	case EChoCtl:
		return "EChoCtl"
	case EchoKe:
		return "EchoKe"
	case PendIn:
		return "PendIn"
	case Opost:
		return "Opost"
	case OlCuc:
		return "OlCuc"
	case ONlCr:
		return "ONlCr"
	case OcrNl:
		return "OcrNl"
	case OnoCr:
		return "OnoCr"
	case OnlRet:
		return "OnlRet"
	case Cs7:
		return "Cs7"
	case Cs8:
		return "Cs8"
	case ParEnb:
		return "ParEnb"
	case ParOdd:
		return "ParOdd"
	case TtyOpInputSpeed:
		return "TtyOpInputSpeed"
	case TtyOpOutputSpeed:
		return "TtyOpOutputSpeed"
	default:
		return fmt.Sprintf("Unknown(%d)", tm)
	}
}
