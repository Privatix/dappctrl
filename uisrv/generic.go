package uisrv

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
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
	OrderingSQL  string
	Paginated    bool
}

// paginatedReply is a format of paginated reply.
type paginatedReply struct {
	Items   []reform.Struct `json:"items"`
	Current int             `json:"current"`
	Pages   int             `json:"pages"`
}

func (s *Server) newPaginatedReply(conf *getConf, tail string,
	args []interface{}, page, perPage int) (*paginatedReply, error) {
	ret := &paginatedReply{
		Current: page,
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`,
		conf.View.Name(), tail)
	row := s.db.QueryRow(query, args...)
	var count int
	if err := row.Scan(&count); err != nil {
		return nil, err
	}

	ret.Pages = int(math.Ceil(
		float64(count) / float64(perPage)))

	return ret, nil
}

func (s *Server) formatConditions(r *http.Request, conf *getConf) (conds []string, args []interface{}) {
	placei := 1

	for _, param := range conf.Params {
		op := "="
		if param.Op != "" {
			op = param.Op
		}

		val := r.FormValue(param.Name)
		if val == "" {
			continue
		}

		var ph string
		if op == "in" {
			subvals := strings.Split(val, ",")
			for _, subval := range subvals {
				args = append(args, subval)
			}

			phs := s.db.Placeholders(placei, len(subvals))
			placei += len(subvals)
			ph = "(" + strings.Join(phs, ",") + ")"
		} else if op == "like" {
			args = append(args, "%"+val+"%")
			ph = s.db.Placeholder(placei)
			placei++
		} else {
			args = append(args, val)
			ph = s.db.Placeholder(placei)
			placei++
		}

		cond := fmt.Sprintf("%s %s %s", param.Field, op, ph)
		conds = append(conds, cond)
	}

	return conds, args
}

// handleGetResources select and returns records.
func (s *Server) handleGetResources(w http.ResponseWriter,
	r *http.Request, conf *getConf) {
	logger := s.logger.Add("method", "handleGetResources",
		"table", conf.View.Name())

	conds, args := s.formatConditions(r, conf)
	if conf.FilteringSQL != "" {
		conds = append(conds, conf.FilteringSQL)
	}

	var filtering, limitOffset string
	if len(conds) > 0 {
		filtering = "WHERE " + strings.Join(conds, " AND ")
	}

	var paginatedItems *paginatedReply

	if conf.Paginated {
		page, err := strconv.Atoi(r.FormValue("page"))
		if err != nil || page == 0 {
			logger.Warn(fmt.Sprintf("invalid param: %v", err))
			s.replyInvalidRequest(logger, w)
			return
		}

		limit, err := strconv.Atoi(r.FormValue("perPage"))
		if err != nil || limit == 0 {
			logger.Warn(fmt.Sprintf("invalid param: %v", err))
			s.replyInvalidRequest(logger, w)
			return
		}

		paginatedItems, err = s.newPaginatedReply(
			conf, filtering, args, page, limit)
		if err != nil {
			logger.Error(
				fmt.Sprintf("failed to get resources: %v", err))
			s.replyUnexpectedErr(logger, w)
			return
		}

		limitOffset = fmt.Sprintf(" LIMIT %d OFFSET %d", limit, (page-1)*limit)
	}

	tail := fmt.Sprintf("%s %s %s", filtering, conf.OrderingSQL, limitOffset)
	records, err := s.db.SelectAllFrom(conf.View, tail, args...)
	if err != nil {
		logger.Warn(fmt.Sprintf("failed to select: %v", err))
		s.replyUnexpectedErr(logger, w)
		return
	}

	if records == nil {
		records = []reform.Struct{}
	}

	if paginatedItems != nil {
		paginatedItems.Items = records
		s.reply(logger, w, paginatedItems)
		return
	}

	s.reply(logger, w, records)
}
