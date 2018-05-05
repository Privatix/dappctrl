package uisrv

import (
	"fmt"
	"net/http"
	"strings"

	reform "gopkg.in/reform.v1"
)

// queryParam is a description of a query param.
type queryParam struct {
	Name  string // in request.
	Field string // column name in db.
	Op    string // comparison operator ({Field} {Op} {Value}, default "=")
}

// getConf is a config for generic get handler.
type getConf struct {
	Params       []queryParam
	View         reform.View
	FilteringSQL string
}

// handleGetResources select and returns records.
func (s *Server) handleGetResources(w http.ResponseWriter,
	r *http.Request, conf *getConf) {
	var tail string
	var eqs []string
	var args []interface{}

	placei := 1

	nextPlaceHolder := func() string {
		ph := s.db.Placeholder(placei)
		placei++
		return ph
	}

	manyPlaceHolders := func(n int) []string {
		phs := s.db.Placeholders(placei, n)
		placei += n
		return phs
	}

	for _, param := range conf.Params {
		op := "="
		if param.Op != "" {
			op = param.Op
		}

		val := r.FormValue(param.Name)
		if val == "" {
			continue
		}

		if op == "in" {
			subvals := strings.Split(val, ",")
			for _, subval := range subvals {
				args = append(args, subval)
			}
			eqs = append(eqs,
				fmt.Sprintf("%s %s (%s)",
					param.Field,
					op,
					strings.Join(manyPlaceHolders(len(subvals)), ",")))
		} else {
			eqs = append(eqs,
				fmt.Sprintf("%s %s %s",
					param.Field,
					op,
					nextPlaceHolder()))
			args = append(args, val)
		}
	}

	if len(args) > 0 {
		tail = "WHERE " + strings.Join(eqs, " AND ")
	}

	if conf.FilteringSQL != "" {
		if tail == "" {
			tail += "WHERE "
		} else {
			tail += " AND "
		}
		tail += conf.FilteringSQL
	}

	records, err := s.db.SelectAllFrom(conf.View, tail, args...)
	if err != nil {
		s.logger.Warn("failed to select: %v", err)
		s.replyUnexpectedErr(w)
		return
	}

	if records == nil {
		records = []reform.Struct{}
	}

	s.reply(w, records)
}
