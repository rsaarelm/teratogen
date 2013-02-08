// +build !windows

package termbox

import "unicode/utf8"
import "bytes"
import "syscall"
import "unsafe"
import "strings"
import "strconv"
import "os"
import "io"

// private API

const (
	t_enter_ca = iota
	t_exit_ca
	t_show_cursor
	t_hide_cursor
	t_clear_screen
	t_sgr0
	t_underline
	t_bold
	t_blink
	t_reverse
	t_enter_keypad
	t_exit_keypad
	t_max_funcs
)

const (
	coord_invalid = -2
	attr_invalid  = Attribute(0xFFFF)
)

type input_event struct {
	data []byte
	err  error
}

var (
	// term specific sequences
	keys  []string
	funcs []string

	// termbox inner state
	orig_tios    syscall_Termios
	back_buffer  cellbuf
	front_buffer cellbuf
	termw        int
	termh        int
	input_mode   = InputEsc
	out          *os.File
	in           int
	lastfg       = attr_invalid
	lastbg       = attr_invalid
	lastx        = coord_invalid
	lasty        = coord_invalid
	cursor_x     = cursor_hidden
	cursor_y     = cursor_hidden
	foreground   = ColorDefault
	background   = ColorDefault
	inbuf        = make([]byte, 0, 64)
	outbuf       bytes.Buffer
	sigwinch     = make(chan os.Signal, 1)
	sigio        = make(chan os.Signal, 1)
	quit         = make(chan int)
	input_comm   = make(chan input_event)
	intbuf       = make([]byte, 0, 16)
)

func write_cursor(x, y int) {
	outbuf.WriteString("\033[")
	outbuf.Write(strconv.AppendUint(intbuf, uint64(y+1), 10))
	outbuf.WriteString(";")
	outbuf.Write(strconv.AppendUint(intbuf, uint64(x+1), 10))
	outbuf.WriteString("H")
}

func write_sgr_fg(a Attribute) {
	outbuf.WriteString("\033[3")
	outbuf.Write(strconv.AppendUint(intbuf, uint64(a-1), 10))
	outbuf.WriteString("m")
}

func write_sgr_bg(a Attribute) {
	outbuf.WriteString("\033[4")
	outbuf.Write(strconv.AppendUint(intbuf, uint64(a-1), 10))
	outbuf.WriteString("m")
}

func write_sgr(fg, bg Attribute) {
	outbuf.WriteString("\033[3")
	outbuf.Write(strconv.AppendUint(intbuf, uint64(fg-1), 10))
	outbuf.WriteString(";4")
	outbuf.Write(strconv.AppendUint(intbuf, uint64(bg-1), 10))
	outbuf.WriteString("m")
}

type winsize struct {
	rows    uint16
	cols    uint16
	xpixels uint16
	ypixels uint16
}

func get_term_size(fd uintptr) (int, int) {
	var sz winsize
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL,
		fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&sz)))
	return int(sz.cols), int(sz.rows)
}

func send_attr(fg, bg Attribute) {
	if fg != lastfg || bg != lastbg {
		outbuf.WriteString(funcs[t_sgr0])
		fgcol := fg & 0x0F
		bgcol := bg & 0x0F
		if fgcol != ColorDefault {
			if bgcol != ColorDefault {
				write_sgr(fgcol, bgcol)
			} else {
				write_sgr_fg(fgcol)
			}
		} else if bgcol != ColorDefault {
			write_sgr_bg(bgcol)
		}

		if fg&AttrBold != 0 {
			outbuf.WriteString(funcs[t_bold])
		}
		if bg&AttrBold != 0 {
			outbuf.WriteString(funcs[t_blink])
		}
		if fg&AttrUnderline != 0 {
			outbuf.WriteString(funcs[t_underline])
		}
		if fg&AttrReverse|bg&AttrReverse != 0 {
			outbuf.WriteString(funcs[t_reverse])
		}

		lastfg, lastbg = fg, bg
	}
}

func send_char(x, y int, ch rune) {
	var buf [8]byte
	n := utf8.EncodeRune(buf[:], ch)
	if x-1 != lastx || y != lasty {
		write_cursor(x, y)
	}
	lastx, lasty = x, y
	outbuf.Write(buf[:n])
}

func flush() error {
	_, err := io.Copy(out, &outbuf)
	outbuf.Reset()
	if err != nil {
		return err
	}
	return nil
}

func send_clear() error {
	send_attr(foreground, background)
	outbuf.WriteString(funcs[t_clear_screen])
	if !is_cursor_hidden(cursor_x, cursor_y) {
		write_cursor(cursor_x, cursor_y)
	}

	// we need to invalidate cursor position too and these two vars are
	// used only for simple cursor positioning optimization, cursor
	// actually may be in the correct place, but we simply discard
	// optimization once and it gives us simple solution for the case when
	// cursor moved
	lastx = coord_invalid
	lasty = coord_invalid

	return flush()
}

func update_size_maybe() error {
	w, h := get_term_size(out.Fd())
	if w != termw || h != termh {
		termw, termh = w, h
		back_buffer.resize(termw, termh)
		front_buffer.resize(termw, termh)
		front_buffer.clear()
		return send_clear()
	}
	return nil
}

func tcsetattr(fd uintptr, termios *syscall_Termios) error {
	r, _, e := syscall.Syscall(syscall.SYS_IOCTL,
		fd, uintptr(syscall_TCSETS), uintptr(unsafe.Pointer(termios)))
	if r != 0 {
		return os.NewSyscallError("SYS_IOCTL", e)
	}
	return nil
}

func tcgetattr(fd uintptr, termios *syscall_Termios) error {
	r, _, e := syscall.Syscall(syscall.SYS_IOCTL,
		fd, uintptr(syscall_TCGETS), uintptr(unsafe.Pointer(termios)))
	if r != 0 {
		return os.NewSyscallError("SYS_IOCTL", e)
	}
	return nil
}

func parse_escape_sequence(event *Event, buf []byte) int {
	bufstr := string(buf)
	for i, key := range keys {
		if strings.HasPrefix(bufstr, key) {
			event.Ch = 0
			event.Key = Key(0xFFFF - i)
			return len(key)
		}
	}
	return 0
}

func extract_event(event *Event) bool {
	if len(inbuf) == 0 {
		return false
	}

	if inbuf[0] == '\033' {
		// possible escape sequence
		n := parse_escape_sequence(event, inbuf)
		if n != 0 {
			copy(inbuf, inbuf[n:])
			inbuf = inbuf[:len(inbuf)-n]
			return true
		}

		// it's not escape sequence, then it's Alt or Esc, check input_mode
		switch input_mode {
		case InputEsc:
			// if we're in escape mode, fill Esc event, pop buffer, return success
			event.Ch = 0
			event.Key = KeyEsc
			event.Mod = 0
			copy(inbuf, inbuf[1:])
			inbuf = inbuf[:len(inbuf)-1]
			return true
		case InputAlt:
			// if we're in alt mode, set Alt modifier to event and redo parsing
			event.Mod = ModAlt
			copy(inbuf, inbuf[1:])
			inbuf = inbuf[:len(inbuf)-1]
			return extract_event(event)
		default:
			panic("unreachable")
		}
	}

	// if we're here, this is not an escape sequence and not an alt sequence
	// so, it's a FUNCTIONAL KEY or a UNICODE character

	// first of all check if it's a functional key
	if Key(inbuf[0]) <= KeySpace || Key(inbuf[0]) == KeyBackspace2 {
		// fill event, pop buffer, return success
		event.Ch = 0
		event.Key = Key(inbuf[0])
		copy(inbuf, inbuf[1:])
		inbuf = inbuf[:len(inbuf)-1]
		return true
	}

	// the only possible option is utf8 rune
	if r, n := utf8.DecodeRune(inbuf); r != utf8.RuneError {
		event.Ch = r
		event.Key = 0
		copy(inbuf, inbuf[n:])
		inbuf = inbuf[:len(inbuf)-n]
		return true
	}

	return false
}

func fcntl(fd int, cmd int, arg int) (val int, err error) {
	r, _, e := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), uintptr(cmd),
		uintptr(arg))
	val = int(r)
	if e != 0 {
		err = e
	}
	return
}
