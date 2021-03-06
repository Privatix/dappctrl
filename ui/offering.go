package ui

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/job"
	"github.com/privatix/dappctrl/messages"
	"github.com/privatix/dappctrl/messages/offer"
	"github.com/privatix/dappctrl/proc/worker"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

// Actions for Agent that change offerings state.
const (
	PublishOffering    = "publish"
	PopupOffering      = "popup"
	DeactivateOffering = "deactivate"
)

// OfferingChangeActions associates an action with a job type.
var OfferingChangeActions = map[string]string{
	PublishOffering:    data.JobAgentPreOfferingMsgBCPublish,
	PopupOffering:      data.JobAgentPreOfferingPopUp,
	DeactivateOffering: data.JobAgentPreOfferingDelete,
}

var offeringCountriesCondition = fmt.Sprintf("%s AND country != ''",
	activeOfferingCondition)

// GetAgentOfferingsResult is result of GetAgentOfferings method.
type GetAgentOfferingsResult struct {
	Items      []data.Offering `json:"items"`
	TotalItems int             `json:"totalItems"`
}

// GetClientOfferingsResultItem is items of GetClientOfferingsResult.
type GetClientOfferingsResultItem struct {
	Offering data.Offering `json:"offering"`
	Rating   uint64        `json:"rating"`
}

// GetClientOfferingsResult is result of GetClientOfferings method.
type GetClientOfferingsResult struct {
	Items      []GetClientOfferingsResultItem `json:"items"`
	TotalItems int                            `json:"totalItems"`
}

// GetClientOfferingsFilterParamsResult is result of
// GetClientOfferingsFilterParams method.
type GetClientOfferingsFilterParamsResult struct {
	Countries []string `json:"countries"`
	MinPrice  uint64   `json:"minPrice"`
	MaxPrice  uint64   `json:"maxPrice"`
	MaxRating uint64   `json:"maxRating"`
}

// AcceptOffering initiates JobClientPreChannelCreate job.
func (h *Handler) AcceptOffering(tkn string, account data.HexString,
	offering string, deposit, gasPrice uint64) (*string, error) {
	logger := h.logger.Add("method", "AcceptOffering",
		"account", account, "offering", offering, "deposit", deposit,
		"gasPrice", gasPrice)

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return nil, ErrAccessDenied
	}

	var acc data.Account
	if err := h.findByColumn(logger, ErrAccountNotFound,
		&acc, "eth_addr", account); err != nil {
		return nil, err
	}

	offer, err := h.findActiveOfferingByID(logger, offering)
	if err != nil {
		return nil, err
	}

	depositSetting, _ := data.GetUint64Setting(h.db, data.SettingClientMinDeposit)
	minDeposit := computeDeposit(offer, depositSetting)

	if deposit == 0 {
		deposit = minDeposit
	} else if deposit < minDeposit {
		logger.Error(ErrDepositTooSmall.Error())
		return nil, ErrDepositTooSmall
	}

	if err := h.pingOffering(logger, offer); err != nil {
		return nil, err
	}

	rid := util.NewUUID()
	jobData := &worker.ClientPreChannelCreateData{Account: acc.ID,
		Offering: offering, GasPrice: gasPrice, Deposit: deposit}
	if err := job.AddWithData(h.queue, nil, data.JobClientPreChannelCreate,
		data.JobChannel, rid, data.JobUser, jobData); err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	return &rid, nil
}

func computeDeposit(offering *data.Offering, configuredMinDeposit uint64) uint64 {
	minDep := data.ComputePrice(offering, offering.MinUnits)
	if configuredMinDeposit > minDep {
		minDep = configuredMinDeposit
	}
	if offering.MaxUnit != nil {
		maxDep := data.ComputePrice(offering, *offering.MaxUnit)
		if maxDep < minDep {
			return maxDep
		}
	}
	return minDep
}

// ChangeOfferingStatus initiates JobAgentPreOfferingMsgBCPublish,
// JobAgentPreOfferingPopUp or JobAgentPreOfferingDelete job,
// depending on a selected action.
func (h *Handler) ChangeOfferingStatus(
	tkn, offering, action string, gasPrice uint64) error {
	logger := h.logger.Add("method", "ChangeOfferingStatus",
		"offering", offering, "action", action, "gasPrice", gasPrice)

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return ErrAccessDenied
	}

	jobType, ok := OfferingChangeActions[action]
	if !ok {
		logger.Warn(ErrBadOfferingStatusAction.Error())
		return ErrBadOfferingStatusAction
	}

	offer := &data.Offering{}
	err := h.findByPrimaryKey(logger, ErrOfferingNotFound, offer, offering)
	if err != nil {
		return err
	}

	jobData := &data.JobPublishData{GasPrice: gasPrice}
	if err := job.AddWithData(h.queue, nil, jobType, data.JobOffering,
		offering, data.JobUser, jobData); err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}

	return nil
}

func (h *Handler) getClientOfferingsConditions(
	agent data.HexString, minUnitPrice, maxUnitPrice uint64,
	country, ipTypes []string) (conditions string, arguments []interface{}) {

	count := 1

	index := func() string {
		current := count
		count++
		return h.db.Placeholder(current)
	}

	join := func(conditions, condition string) string {
		if conditions == "" {
			return condition
		}
		return fmt.Sprintf("%s AND %s", conditions, condition)
	}

	if agent != "" {
		condition := fmt.Sprintf("%s = %s", "agent", index())
		conditions = join(conditions, condition)
		arguments = append(arguments, agent)
	}

	if minUnitPrice > 0 {
		condition := fmt.Sprintf("%s >= %s", "unit_price", index())
		conditions = join(conditions, condition)
		arguments = append(arguments, minUnitPrice)

	}

	if maxUnitPrice > 0 {
		condition := fmt.Sprintf("%s <= %s", "unit_price", index())
		conditions = join(conditions, condition)
		arguments = append(arguments, maxUnitPrice)
	}

	if len(country) != 0 {
		indexes := h.db.Placeholders(count, len(country))
		count = count + len(country)

		condition := fmt.Sprintf("country IN (%s)",
			strings.Join(indexes, ","))
		conditions = join(conditions, condition)

		for _, val := range country {
			arguments = append(arguments, val)
		}
	}

	if len(ipTypes) != 0 {
		indexes := h.db.Placeholders(count, len(ipTypes))
		count = count + len(ipTypes)

		condition := fmt.Sprintf("ip_type IN (%s)",
			strings.Join(indexes, ","))
		conditions = join(conditions, condition)

		for _, val := range ipTypes {
			arguments = append(arguments, val)
		}
	}

	var format string
	if conditions != "" {
		format = "WHERE %s AND %s"
	} else {
		format = "WHERE %s%s"
	}

	conditions = fmt.Sprintf(format, conditions, activeOfferingCondition)
	return conditions, arguments
}

// GetClientOfferings returns active offerings available for a client.
func (h *Handler) GetClientOfferings(tkn string, agent data.HexString,
	minUnitPrice, maxUnitPrice uint64, countries, ipTypes []string,
	offset, limit uint) (*GetClientOfferingsResult, error) {
	logger := h.logger.Add("method", "GetClientOfferings",
		"agent", agent, "minUnitPrice", minUnitPrice,
		"maxUnitPrice", maxUnitPrice, "countries", countries, "offset", offset,
		"limit", limit)

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return nil, ErrAccessDenied
	}

	if minUnitPrice != 0 && maxUnitPrice != 0 &&
		minUnitPrice > maxUnitPrice {
		logger.Error(ErrBadUnitPriceRange.Error())
		return nil, ErrBadUnitPriceRange
	}

	cond, args := h.getClientOfferingsConditions(agent, minUnitPrice,
		maxUnitPrice, countries, ipTypes)

	count, err := h.numberOfObjects(
		logger, data.OfferingTable.Name(), cond, args)
	if err != nil {
		return nil, err
	}

	offsetLimit := h.offsetLimit(offset, limit)

	tail := fmt.Sprintf("%s %s %s",
		cond, activeOfferingSorting, offsetLimit)

	result, err := h.selectAllFrom(
		logger, data.OfferingTable, tail, args...)
	if err != nil {
		return nil, err
	}

	items := make([]GetClientOfferingsResultItem, len(result))
	ratings := make(map[data.HexString]uint64)
	for k, v := range result {
		items[k].Offering = *v.(*data.Offering)
		if val, ok := ratings[items[k].Offering.Agent]; ok {
			items[k].Rating = val
		} else {
			var rating data.Rating
			if err := h.db.FindByPrimaryKeyTo(&rating, items[k].Offering.Agent); err != nil {
				if err != sql.ErrNoRows {
					logger.Error(err.Error())
					return nil, ErrInternal
				}
			}
			items[k].Rating = rating.Val
		}
	}

	return &GetClientOfferingsResult{items, count}, nil
}

func (h *Handler) getAgentOfferingsConditions(
	product string, statuses []string) (conditions string, args []interface{}) {
	index := 1

	if product != "" {
		conditions = fmt.Sprintf(
			"product = %s", h.db.Placeholder(index))
		args = append(args, product)
		index++
	}

	if length := len(statuses); length != 0 {

		placeholders := h.db.Placeholders(index, length)

		condition := fmt.Sprintf(
			"status IN ( %s )", strings.Join(placeholders, ", "))
		for _, v := range statuses {
			args = append(args, v)
		}

		index += length

		if conditions == "" {
			conditions = condition
		} else {
			conditions = fmt.Sprintf(
				"%s AND %s", conditions, condition)
		}
	}

	condition := `
		agent IN (SELECT eth_addr FROM accounts)
			AND (SELECT in_use FROM accounts WHERE eth_addr = agent)`

	if conditions == "" {
		conditions = "WHERE " + condition
	} else {
		conditions = fmt.Sprintf("WHERE %s AND %s",
			conditions, condition)
	}

	return conditions, args
}

// GetAgentOfferings returns active offerings available for a agent.
func (h *Handler) GetAgentOfferings(tkn, product string, statuses []string,
	offset, limit uint) (*GetAgentOfferingsResult, error) {
	logger := h.logger.Add("method", "GetAgentOfferings",
		"product", product, "status", statuses)

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return nil, ErrAccessDenied
	}

	conditions, args := h.getAgentOfferingsConditions(product, statuses)

	count, err := h.numberOfObjects(
		logger, data.OfferingTable.Name(), conditions, args)
	if err != nil {
		return nil, err
	}

	offsetLimit := h.offsetLimit(offset, limit)

	sorting := `ORDER BY block_number_updated DESC`

	tail := fmt.Sprintf("%s %s %s", conditions, sorting, offsetLimit)

	result, err := h.selectAllFrom(
		logger, data.OfferingTable, tail, args...)
	if err != nil {
		return nil, err
	}

	offerings := make([]data.Offering, len(result))

	for k, v := range result {
		offerings[k] = *v.(*data.Offering)
	}

	return &GetAgentOfferingsResult{offerings, count}, nil
}

// setOfferingHash computes and sets values for raw msg and hash fields.
func (h *Handler) setOfferingHash(logger log.Logger, offering *data.Offering,
	template *data.Template, agent *data.Account) error {
	handleErr := func(err error) error {
		logger.Error(err.Error())
		return ErrInternal
	}
	msg := offer.OfferingMessage(agent, template, offering)

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return handleErr(err)
	}

	agentKey, err := h.pwdStorage.GetKey(agent)
	if err != nil {
		return handleErr(err)
	}

	packed, err := messages.PackWithSignature(msgBytes, agentKey)
	if err != nil {
		return handleErr(err)
	}

	offering.RawMsg = data.FromBytes(packed)

	hashBytes := common.BytesToHash(crypto.Keccak256(packed))

	offering.Hash = data.HexFromBytes(hashBytes.Bytes())

	return nil
}

// fillOffering fills offerings nonce, status, hash and signature.
func (h *Handler) fillOffering(
	logger log.Logger, offering *data.Offering) error {
	agent := &data.Account{}
	// TODO: This is definitely wrong, should be:
	// `h.findByColumn(..., "eth_addr", offering.Agent)`
	if err := h.findByPrimaryKey(logger,
		ErrAccountNotFound, agent, string(offering.Agent)); err != nil {
		return err
	}

	template := &data.Template{}
	if err := h.findByPrimaryKey(logger,
		ErrTemplateNotFound, template, offering.Template); err != nil {
		return err
	}

	offering.ID = util.NewUUID()
	offering.Status = data.OfferEmpty
	offering.Agent = agent.EthAddr
	offering.BlockNumberUpdated = 1
	offering.CurrentSupply = offering.Supply
	// TODO: remove once prepaid is implemented.
	offering.BillingType = data.BillingPostpaid

	return h.setOfferingHash(logger, offering, template, agent)
}

func (h *Handler) prepareOffering(
	logger log.Logger, offering *data.Offering) error {
	if offering.UnitType != data.UnitScalar &&
		offering.UnitType != data.UnitSeconds {
		logger.Error(ErrBadUnitType.Error())
		return ErrBadUnitType
	}

	if offering.BillingType != data.BillingPrepaid &&
		offering.BillingType != data.BillingPostpaid {
		logger.Error(ErrBillingType.Error())
		return ErrBillingType
	}
	return h.fillOffering(logger, offering)
}

// UpdateOffering updates an offering.
func (h *Handler) UpdateOffering(tkn string, offering *data.Offering) error {
	logger := h.logger.Add(
		"method", "UpdateOffering", "offering", offering)

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return ErrAccessDenied
	}

	err := h.findByPrimaryKey(
		logger, ErrOfferingNotFound, &data.Offering{}, offering.ID)
	if err != nil {
		return err
	}

	err = update(logger, h.db.Querier, offering)
	if err != nil {
		return err
	}

	return nil
}

// CreateOffering creates an offering.
func (h *Handler) CreateOffering(tkn string,
	offering *data.Offering) (*string, error) {
	logger := h.logger.Add(
		"method", "CreateOffering", "offering", offering)

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return nil, ErrAccessDenied
	}

	err := h.prepareOffering(logger, offering)
	if err != nil {
		return nil, err
	}

	err = insert(logger, h.db.Querier, offering)
	if err != nil {
		return nil, err
	}

	return &offering.ID, nil
}

func (h *Handler) offeringCountries(logger log.Logger) ([]string, error) {
	query := `SELECT country, COUNT(country)
		    FROM offerings WHERE %s
		   GROUP BY country
		   ORDER BY count DESC`

	tail := fmt.Sprintf(query, offeringCountriesCondition)

	rows, err := h.db.Query(tail)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}
	defer rows.Close()

	countries := make([]string, 0)

	for rows.Next() {
		var country string
		var count int64

		if err := rows.Scan(&country, &count); err != nil {
			logger.Error(err.Error())
			return nil, ErrInternal
		}
		countries = append(countries, country)
	}

	if err := rows.Err(); err != nil {
		logger.Error(err.Error())
	}

	return countries, nil
}

func (h *Handler) offeringsMinMaxPrice(
	logger log.Logger) (min uint64, max uint64, err error) {
	query := `SELECT COALESCE(MIN(unit_price),0), COALESCE(MAX(unit_price),0)
		    FROM offerings WHERE %s`

	tail := fmt.Sprintf(query, offeringCountriesCondition)

	if err := h.db.QueryRow(tail).Scan(&min, &max); err != nil {
		logger.Error(err.Error())
		return 0, 0, ErrInternal
	}
	return min, max, nil
}

func (h *Handler) maxRating(logger log.Logger) (uint64, error) {
	query := `SELECT COALESCE(MAX(val), 0) FROM ratings`
	var ret uint64
	if err := h.db.QueryRow(query).Scan(&ret); err != nil {
		logger.Error(err.Error())
		return 0, err
	}
	return ret, nil
}

// GetClientOfferingsFilterParams returns offerings filter parameters for client.
func (h *Handler) GetClientOfferingsFilterParams(
	tkn string) (*GetClientOfferingsFilterParamsResult, error) {
	logger := h.logger.Add("method", "GetClientOfferingsFilterParams")

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return nil, ErrAccessDenied
	}

	countries, err := h.offeringCountries(logger)
	if err != nil {
		return nil, err
	}

	min, max, err := h.offeringsMinMaxPrice(logger)
	if err != nil {
		return nil, err
	}

	maxRating, err := h.maxRating(logger)
	if err != nil {
		return nil, err
	}

	return &GetClientOfferingsFilterParamsResult{countries, min, max, maxRating}, nil
}

// PingOfferings given offerings ids pings each of them and returns result of the test.
func (h *Handler) PingOfferings(tkn string, ids []string) (map[string]bool, error) {
	logger := h.logger.Add("method", "PingOfferings", "ids", ids)

	if !h.token.Check(tkn) {
		logger.Warn("access denied")
		return nil, ErrAccessDenied
	}

	now := time.Now()
	ret := make(map[string]bool)

	wg := new(sync.WaitGroup)
	wg.Add(len(ids))
	for _, id := range ids {
		offering, err := h.findActiveOfferingByID(logger, id)
		if err != nil {
			return nil, err
		}
		go func(offering *data.Offering) {
			err := h.pingOffering(logger, offering)
			if err == nil {
				ret[offering.ID] = true
				offering.SOMCSuccessPing = &now
				err = update(logger, h.db.Querier, offering)
				if err != nil {
					logger.Warn(err.Error())
				}
			} else {
				ret[offering.ID] = false
				logger.Debug(err.Error())
			}
			wg.Done()
		}(offering)
	}
	wg.Wait()
	return ret, nil
}

func (h *Handler) pingOffering(logger log.Logger, offering *data.Offering) error {
	client, err := h.somcClientBuilder.NewClient(offering.SOMCType, offering.SOMCData)
	if err != nil {
		logger.Error(err.Error())
		return ErrSOMCIsNotAvailable
	}
	err = client.Ping()
	if err != nil {
		logger.Warn(err.Error())
		return ErrSOMCIsNotAvailable
	}
	return err
}
