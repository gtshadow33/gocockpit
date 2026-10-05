package terminal

import (
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"sync"
	"syscall"

	"github.com/creack/pty"
)

type Terminal struct {
	Cmd  *exec.Cmd
	PTY  *os.File
	once sync.Once
}

// Start lanza una shell de login como el usuario indicado, dentro de una PTY.
// Requiere que GoCockpit corra como root (o con CAP_SETUID, CAP_SETGID).
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

	// Grupos suplementarios del usuario (wheel, docker...).
	// Sin esto Go los vacía y la shell no podría usar sudo ni docker.
	groupIDs, err := u.GroupIds()
	if err != nil {
		return nil, err
	}

	groups := make([]uint32, 0, len(groupIDs))

	for _, g := range groupIDs {
		n, err := strconv.ParseUint(g, 10, 32)
		if err == nil {
			groups = append(groups, uint32(n))
		}
	}

	cmd := exec.Command("/bin/bash", "-l")

	cmd.Dir = u.HomeDir

	// Si el home no existe, se arranca en la raíz.
	if _, err := os.Stat(u.HomeDir); err != nil {
		cmd.Dir = "/"
	}

	// Entorno limpio: no se hereda el de root / el del servicio,
	// así no se filtran variables de GoCockpit al usuario.
	cmd.Env = []string{
		"HOME=" + u.HomeDir,
		"USER=" + u.Username,
		"LOGNAME=" + u.Username,
		"SHELL=/bin/bash",
		"TERM=xterm-256color",
		"LANG=C.UTF-8",
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid:    uint32(uid),
			Gid:    uint32(gid),
			Groups: groups,
		},
		// Sesión propia: el PID de bash pasa a ser el ID del grupo de
		// procesos, necesario para matar también a sus hijos.
		Setsid:  true,
		Setctty: true,
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

// Close cierra la PTY, mata todo el grupo de procesos de la shell
// y recoge el proceso para que no quede zombie. Es seguro llamarla
// varias veces.
func (t *Terminal) Close() error {

	t.once.Do(func() {

		if t.PTY != nil {
			t.PTY.Close()
		}

		if t.Cmd != nil && t.Cmd.Process != nil {

			// PID negativo = todo el grupo (bash + vim, htop, sleep...).
			syscall.Kill(-t.Cmd.Process.Pid, syscall.SIGKILL)

			// Wait recoge el proceso; en goroutine para no bloquear.
			go t.Cmd.Wait()
		}
	})

	return nil
}