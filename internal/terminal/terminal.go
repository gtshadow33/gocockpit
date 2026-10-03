package terminal

import (
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"github.com/creack/pty"
)

type Terminal struct {
	Cmd *exec.Cmd
	PTY *os.File
}

func Start(username string) (*Terminal, error) {

	u, err := user.Lookup(username)
	if err != nil {
		return nil, err
	}

	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return nil, err
	}

	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("/bin/bash")

	cmd.Dir = u.HomeDir

	cmd.Env = append(
		os.Environ(),
		"HOME="+u.HomeDir,
		"USER="+u.Username,
		"LOGNAME="+u.Username,
		"SHELL=/bin/bash",
		"TERM=xterm-256color",
	)

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	return &Terminal{
		Cmd: cmd,
		PTY: ptmx,
	}, nil
}

func (t *Terminal) Close() error {

	if t.PTY != nil {
		t.PTY.Close()
	}

	if t.Cmd.Process != nil {
		return t.Cmd.Process.Kill()
	}

	return nil
}