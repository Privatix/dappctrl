package ui

import (
	"fmt"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util/log"
)

// GetLogsResult is result of GetLogs method.
type GetLogsResult struct {
	Items      []data.LogEvent `json:"items"`
	TotalItems int             `json:"totalItems"`
}

type getLogsArgs struct {
	level      string
	dateFrom   string
	dateTo     string
	searchText string
}

func (h *Handler) getTotalLogEvents(logger log.Logger, conditions string,
	arguments []interface{}) (count int, err error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`,
		data.LogEventView.Name(), conditions)

	err = h.db.QueryRow(query, arguments...).Scan(&count)
	if err != nil {
		logger.Error(err.Error())
		return 0, ErrInternal
	}

	return count, err
}

func (h *Handler) getLogsConditions(
	args *getLogsArgs) (conditions string, arguments []interface{}) {

	var contextSearchSQL string

	if args.searchText != "" {
		contextSearchSQL = fmt.Sprintf("to_tsvector('english',"+
			" context) @@ to_tsquery('%s:*')", args.searchText)
	}

	count := 1

	index := func() string {
		current := count
		count++
		return h.db.Placeholder(current)
	}

	join := func(conditions, operator, condition string) string {
		if conditions == "" {
			return condition
		}
		return fmt.Sprintf("%s %s %s", conditions, operator, condition)
	}

	if args.level != "" {
		condition := fmt.Sprintf("%s = %s", "level", index())
		conditions = join(conditions, "AND", condition)
		arguments = append(arguments, args.level)
	}

	if args.dateFrom != "" {
		condition := fmt.Sprintf("%s >= %s", "time", index())
		conditions = join(conditions, "AND", condition)
		arguments = append(arguments, args.dateFrom)
	}

	if args.dateTo != "" {
		condition := fmt.Sprintf("%s < %s", "time", index())
		conditions = join(conditions, "AND", condition)
		arguments = append(arguments, args.dateTo)
	}

	if contextSearchSQL != "" {
		conditions = join(conditions, "OR", contextSearchSQL)
	}

	conditions = "WHERE " + conditions

	return conditions, arguments
}

func (h *Handler) getLogs(logger log.Logger, conditions string,
	arguments []interface{}, offset, limit uint) ([]data.LogEvent, error) {

	var limitCondition string

	if limit != 0 {
		limitCondition = fmt.Sprintf(" LIMIT %d ", limit)
	}

	offsetCondition := fmt.Sprintf(" OFFSET %d", offset)

	conditions = conditions + " ORDER BY time DESC " +
		limitCondition + offsetCondition

	records, err := h.db.SelectAllFrom(
		data.LogEventView, conditions, arguments...)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	var result []data.LogEvent
	for k := range records {
		result = append(result, *records[k].(*data.LogEvent))
	}
	return result, nil
}

// GetLogs returns back end log, paginated.
func (h *Handler) GetLogs(password string, offset, limit uint,
	searchText, level, dateFrom, dateTo string) (*GetLogsResult, error) {
	logger := h.logger.Add("method", "GetLogs", "offset", offset,
		"limit", limit, "searchText", searchText, "level", level,
		"dateFrom", dateFrom, "dateTo", dateTo)

	err := h.checkPassword(logger, password)
	if err != nil {
		return nil, err
	}

	args := &getLogsArgs{
		level:      level,
		dateTo:     dateTo,
		dateFrom:   dateFrom,
		searchText: searchText,
	}

	conditions, arguments := h.getLogsConditions(args)
	totalItems, err := h.getTotalLogEvents(logger, conditions, arguments)
	if err != nil {
		return nil, err
	}

	result, err := h.getLogs(logger, conditions, arguments, offset, limit)
	if err != nil {
		return nil, err
	}

	return &GetLogsResult{result, totalItems}, err
}
