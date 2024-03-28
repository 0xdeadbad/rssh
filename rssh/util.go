package rssh

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/creack/pty"
	"golang.org/x/sys/unix"
	"log"
	"os"
)

func generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	log.Println("Private Key generated")
	return privateKey, nil
}

func GetPty(ptyReq *PtyRequest) (*os.File, *os.File, error) {
	ppty, tty, err := pty.Open()
	if err != nil {
		return nil, nil, err
	}

	/*tmios := unix.Termios{}

	for k, v := range ptyReq.Modes {
		switch k {
		case TtyOpEnd:
			break
		case VIntr, VQuit, VErase, VKill, VEof, VEol, VEol2, VStart, VStop, VSusp, VDsusp, VReprint, VWErase, VLNExt, VFLush, VSwtch, VStatus, VDiscard:
			tmios.Cc[k] = 0
		case IgnPar, ParMrk, InpCk, IStrip, InlCr, IgnCr, ICrnl, IUclc, IXon, IXany, IXoff, IMaxbel, IUTF8:
			tmios.Iflag |= uint32(k)
		case ISig, ICanon, XCase, Echo, EchoE, EchoK, EchoNL, EChoCtl, EchoKe, NoFlsh, ToStop, PendIn, IExten:
			tmios.Lflag |= uint32(k)
		case Opost, OlCuc, ONlCr, OcrNl, OnoCr, OnlRet:
			tmios.Oflag |= uint32(k)
		case ParEnb, ParOdd, Cs7, Cs8:
			tmios.Cflag |= uint32(k)
		case TtyOpInputSpeed, TtyOpOutputSpeed:
			tmios.Ispeed = uint32(k)
			tmios.Ospeed = uint32(k)
		default:
			log.Printf("Unknown mode %d:%d\n", k, v)
			break
		}
	}

	err = unix.IoctlSetTermios(int(ppty.Fd()), unix.TCSETA, &tmios)
	if err != nil {
		return nil, nil, err
	}*/

	wsz := unix.Winsize{
		Row:    uint16(ptyReq.Rows),
		Col:    uint16(ptyReq.Columns),
		Xpixel: uint16(ptyReq.Width),
		Ypixel: uint16(ptyReq.Height),
	}

	err = unix.IoctlSetWinsize(int(ppty.Fd()), unix.TIOCSWINSZ, &wsz)
	if err != nil {
		return nil, nil, err
	}

	return ppty, tty, nil
}
