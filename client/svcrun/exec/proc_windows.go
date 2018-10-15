package exec

import (
	"io"
	"os"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/alexbrainman/ps/winapi"
)

//go:generate go run $GOROOT/src/syscall/mksyscall_windows.go -output proc_gen.go proc_windows.go

//sys resumeThread(thread syscall.Handle) (err error) = kernel32.ResumeThread

// Process is a process, which allows to kill itself together with its children.
type Process struct {
	Stderr io.ReadCloser
	job    syscall.Handle
	handle syscall.Handle
}

// StartProcess starts a new process with a given name and arguments.
func StartProcess(name string, arg ...string) (*Process, error) {
	var in, out, job syscall.Handle
	var info winapi.JOBOBJECT_EXTENDED_LIMIT_INFORMATION
	var attrs syscall.ProcAttr
	var pi *syscall.ProcessInformation

	err := syscall.CreatePipe(&in, &out, nil, 0)
	if err != nil {
		goto failed
	}

	job, err = winapi.CreateJobObject(nil, nil)
	if err != nil {
		goto failed
	}

	info = winapi.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: winapi.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: winapi.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}
	err = winapi.SetInformationJobObject(job,
		winapi.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)), uint32(unsafe.Sizeof(info)))
	if err != nil {
		goto failed
	}

	const createSuspended = 0x00000004
	attrs = syscall.ProcAttr{
		Files: []uintptr{0, 0, uintptr(out)},
		Sys: &syscall.SysProcAttr{
			CreationFlags: createSuspended,
		},
	}

	pi, err = start(name, append([]string{name}, arg...), &attrs)
	if err != nil {
		goto failed
	}

	err = winapi.AssignProcessToJobObject(job, pi.Process)
	if err != nil {
		goto failed
	}

	err = resumeThread(pi.Thread)
	syscall.CloseHandle(pi.Thread)
	if err != nil {
		goto failed
	}

	syscall.CloseHandle(out)

	return &Process{os.NewFile(uintptr(in), ""), job, pi.Process}, nil

failed:
	syscall.CloseHandle(job)
	syscall.CloseHandle(out)
	syscall.CloseHandle(in)

	return nil, err
}

// Close closes the instance without killing the underlying process.
func (p *Process) Close() {
	p.Stderr.Close()
	syscall.CloseHandle(p.handle)
	syscall.CloseHandle(p.job)
}

// Wait waits for the process to exit.
func (p *Process) Wait() error {
	_, err := syscall.WaitForSingleObject(p.handle, syscall.INFINITE)
	return err
}

// Kill kills the process.
func (p *Process) Kill() error {
	return syscall.TerminateProcess(p.handle, 1)
}

// The following code of start() function is a slightly modified version
// of syscall.StartProcess() implementation for Windows.

func start(argv0 string, argv []string, attr *syscall.ProcAttr) (pi *syscall.ProcessInformation, err error) {
	if len(argv0) == 0 {
		return nil, syscall.EWINDOWS
	}
	if attr == nil {
		attr = &zeroProcAttr
	}
	sys := attr.Sys
	if sys == nil {
		sys = &zeroSysProcAttr
	}

	if len(attr.Files) > 3 {
		return nil, syscall.EWINDOWS
	}
	if len(attr.Files) < 3 {
		return nil, syscall.EINVAL
	}

	if len(attr.Dir) != 0 {
		// StartProcess assumes that argv0 is relative to attr.Dir,
		// because it implies Chdir(attr.Dir) before executing argv0.
		// Windows CreateProcess assumes the opposite: it looks for
		// argv0 relative to the current directory, and, only once the new
		// process is started, it does Chdir(attr.Dir). We are adjusting
		// for that difference here by making argv0 absolute.
		var err error
		argv0, err = joinExeDirAndFName(attr.Dir, argv0)
		if err != nil {
			return nil, err
		}
	}
	argv0p, err := syscall.UTF16PtrFromString(argv0)
	if err != nil {
		return nil, err
	}

	var cmdline string
	// Windows CreateProcess takes the command line as a single string:
	// use attr.CmdLine if set, else build the command line by escaping
	// and joining each argument with spaces
	if sys.CmdLine != "" {
		cmdline = sys.CmdLine
	} else {
		cmdline = makeCmdLine(argv)
	}

	var argvp *uint16
	if len(cmdline) != 0 {
		argvp, err = syscall.UTF16PtrFromString(cmdline)
		if err != nil {
			return nil, err
		}
	}

	var dirp *uint16
	if len(attr.Dir) != 0 {
		dirp, err = syscall.UTF16PtrFromString(attr.Dir)
		if err != nil {
			return nil, err
		}
	}

	// Acquire the fork lock so that no other threads
	// create new fds that are not yet close-on-exec
	// before we fork.
	syscall.ForkLock.Lock()
	defer syscall.ForkLock.Unlock()

	p, _ := syscall.GetCurrentProcess()
	fd := make([]syscall.Handle, len(attr.Files))
	for i := range attr.Files {
		if attr.Files[i] > 0 {
			err := syscall.DuplicateHandle(p, syscall.Handle(attr.Files[i]), p, &fd[i], 0, true, syscall.DUPLICATE_SAME_ACCESS)
			if err != nil {
				return nil, err
			}
			defer syscall.CloseHandle(syscall.Handle(fd[i]))
		}
	}
	si := new(syscall.StartupInfo)
	si.Cb = uint32(unsafe.Sizeof(*si))
	si.Flags = syscall.STARTF_USESTDHANDLES
	if sys.HideWindow {
		si.Flags |= syscall.STARTF_USESHOWWINDOW
		si.ShowWindow = syscall.SW_HIDE
	}
	si.StdInput = fd[0]
	si.StdOutput = fd[1]
	si.StdErr = fd[2]

	pi = new(syscall.ProcessInformation)

	flags := sys.CreationFlags | syscall.CREATE_UNICODE_ENVIRONMENT
	if sys.Token != 0 {
		err = syscall.CreateProcessAsUser(sys.Token, argv0p, argvp, nil, nil, true, flags, createEnvBlock(attr.Env), dirp, si, pi)
	} else {
		err = syscall.CreateProcess(argv0p, argvp, nil, nil, true, flags, createEnvBlock(attr.Env), dirp, si, pi)
	}
	if err != nil {
		return nil, err
	}

	return pi, nil
}

func makeCmdLine(args []string) string {
	var s string
	for _, v := range args {
		if s != "" {
			s += " "
		}
		s += syscall.EscapeArg(v)
	}
	return s
}

func joinExeDirAndFName(dir, p string) (name string, err error) {
	if len(p) == 0 {
		return "", syscall.EINVAL
	}
	if len(p) > 2 && isSlash(p[0]) && isSlash(p[1]) {
		// \\server\share\path form
		return p, nil
	}
	if len(p) > 1 && p[1] == ':' {
		// has drive letter
		if len(p) == 2 {
			return "", syscall.EINVAL
		}
		if isSlash(p[2]) {
			return p, nil
		}
		d, err := normalizeDir(dir)
		if err != nil {
			return "", err
		}
		if volToUpper(int(p[0])) == volToUpper(int(d[0])) {
			return syscall.FullPath(d + "\\" + p[2:])
		}
		return syscall.FullPath(p)
	}
	// no drive letter
	d, err := normalizeDir(dir)
	if err != nil {
		return "", err
	}
	if isSlash(p[0]) {
		return syscall.FullPath(d[:2] + p)
	}
	return syscall.FullPath(d + "\\" + p)
}

var zeroProcAttr syscall.ProcAttr
var zeroSysProcAttr syscall.SysProcAttr

func createEnvBlock(envv []string) *uint16 {
	if len(envv) == 0 {
		return &utf16.Encode([]rune("\x00\x00"))[0]
	}
	length := 0
	for _, s := range envv {
		length += len(s) + 1
	}
	length++

	b := make([]byte, length)
	i := 0
	for _, s := range envv {
		l := len(s)
		copy(b[i:i+l], []byte(s))
		copy(b[i+l:i+l+1], []byte{0})
		i = i + l + 1
	}
	copy(b[i:i+1], []byte{0})

	return &utf16.Encode([]rune(string(b)))[0]
}

func isSlash(c uint8) bool {
	return c == '\\' || c == '/'
}

func normalizeDir(dir string) (name string, err error) {
	ndir, err := syscall.FullPath(dir)
	if err != nil {
		return "", err
	}
	if len(ndir) > 2 && isSlash(ndir[0]) && isSlash(ndir[1]) {
		// dir cannot have \\server\share\path form
		return "", syscall.EINVAL
	}
	return ndir, nil
}

func volToUpper(ch int) int {
	if 'a' <= ch && ch <= 'z' {
		ch += 'A' - 'a'
	}
	return ch
}
