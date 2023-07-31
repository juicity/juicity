package dialer

import (
	"fmt"
	"syscall"

	"github.com/mzz2017/softwind/netproxy"
)

func init() {
	soMark := netproxy.SoMark
	netproxy.SoMark = func(fd, mark int) error {
		if protectPath == "" {
			return soMark(fd, mark)
		}
		if err := protect(fd, protectPath); err != nil {
			return fmt.Errorf("protect failed: %w", err)
		}
		return nil
	}
	soMarkControl := netproxy.SoMarkControl
	netproxy.SoMarkControl = func(c syscall.RawConn, mark int) error {
		if protectPath == "" {
			return soMarkControl(c, mark)
		}
		var sockOptErr error
		controlErr := c.Control(func(fd uintptr) {
			err := protect(int(fd), protectPath)
			if err != nil {
				sockOptErr = fmt.Errorf("error setting SO_MARK socket option: %w", err)
			}
		})
		if controlErr != nil {
			return fmt.Errorf("error invoking socket control function: %w", controlErr)
		}
		return sockOptErr
	}
}

func protect(fd int, unixPath string) error {
	if fd <= 0 {
		return nil
	}

	socket, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(socket)

	syscall.SetsockoptTimeval(socket, syscall.SOL_SOCKET, syscall.SO_RCVTIMEO, &syscall.Timeval{Sec: 3})
	syscall.SetsockoptTimeval(socket, syscall.SOL_SOCKET, syscall.SO_SNDTIMEO, &syscall.Timeval{Sec: 3})

	err = syscall.Connect(socket, &syscall.SockaddrUnix{Name: unixPath})
	if err != nil {
		return err
	}

	err = syscall.Sendmsg(socket, nil, syscall.UnixRights(fd), nil, 0)
	if err != nil {
		return err
	}

	dummy := []byte{1}
	n, err := syscall.Read(socket, dummy)
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("protect failed")
	}
	return nil
}
