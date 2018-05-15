package execsrv

// TODO: This temporal package is intended for DevOps testing only, so please
// remove it when not needed anymore.

import (
	"net/http"
	"os/exec"

	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/srv"
)

// Server is a server which allows to launch local processes remotely.
type Server struct {
	*srv.Server
}

// NewServer creates a new launching server.
func NewServer(logger *util.Logger) *Server {
	conf := srv.NewConfig()
	conf.Addr = "0.0.0.0:1234"

	s := &Server{
		Server: srv.NewServer(conf, logger),
	}

	s.HandleFunc("/", s.handle)

	return s
}

type args struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}

type result struct {
	PID int `json:"pid"`
}

func (s *Server) handle(
	w http.ResponseWriter, r *http.Request, ctx *srv.Context) {
	var args args
	if !s.ParseRequest(w, r, &args) {
		return
	}

	cmd := exec.Command(args.Name, args.Args...)

	if err := cmd.Start(); err != nil {
		s.RespondError(w, &srv.Error{
			Code:    srv.ErrCodeMax + 1,
			Message: err.Error(),
		})
		return
	}

	s.RespondResult(w, result{cmd.Process.Pid})
}
