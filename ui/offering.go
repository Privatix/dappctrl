package ui

import (
	"encoding/json"
	"fmt"
	"strings"

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

// AcceptOffering initiates JobClientPreChannelCreate job.
func (h *Handler) AcceptOffering(password, account, offering string,
	deposit, gasPrice uint64) (*string, error) {
	logger := h.logger.Add("method", "AcceptOffering",
		"account", account, "offering", offering,
		"deposit", deposit, "gasPrice", gasPrice)

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	var acc data.Account
	if err := h.findByPrimaryKey(
		logger, ErrAccountNotFound, &acc, account); err != nil {
		return nil, err
	}

	offer, err := h.findActiveOfferingByID(logger, offering)
	if err != nil {
		return nil, err
	}

	minDeposit := data.MinDeposit(offer)

	if deposit == 0 {
		deposit = minDeposit
	} else if deposit < minDeposit {
		logger.Error(ErrDepositTooSmall.Error())
		return nil, ErrDepositTooSmall
	}

	rid := util.NewUUID()
	jobData := &worker.ClientPreChannelCreateData{Account: account,
		Offering: offering, GasPrice: gasPrice, Deposit: deposit}
	if err := job.AddWithData(h.queue, nil, data.JobClientPreChannelCreate,
		data.JobChannel, rid, data.JobUser, jobData); err != nil {
		logger.Error(err.Error())
		return nil, ErrInternal
	}

	return &rid, nil
}

// ChangeOfferingStatus initiates JobAgentPreOfferingMsgBCPublish,
// JobAgentPreOfferingPopUp or JobAgentPreOfferingDelete job,
// depending on a selected action.
func (h *Handler) ChangeOfferingStatus(
	password, offering, action string, gasPrice uint64) error {
	logger := h.logger.Add("method", "ChangeOfferingStatus",
		"offering", offering, "action", action, "gasPrice", gasPrice)

	if err := h.checkPassword(logger, password); err != nil {
		return err
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
	agent string, minUnitPrice, maxUnitPrice uint64,
	country []string) (conditions string, arguments []interface{}) {

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
func (h *Handler) GetClientOfferings(
	password, agent string, minUnitPrice, maxUnitPrice uint64,
	countries []string) ([]data.Offering, error) {
	logger := h.logger.Add("method", "GetClientOfferings",
		"agent", agent, "minUnitPrice", minUnitPrice,
		"maxUnitPrice", maxUnitPrice, "countries", countries)

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	if minUnitPrice != 0 && maxUnitPrice != 0 &&
		minUnitPrice > maxUnitPrice {
		logger.Error(ErrBadUnitPriceRange.Error())
		return nil, ErrBadUnitPriceRange
	}

	cond, args := h.getClientOfferingsConditions(
		agent, minUnitPrice, maxUnitPrice, countries)

	objects, err := h.selectAllFrom(
		logger, data.OfferingTable, cond, args...)
	if err != nil {
		return nil, err
	}

	var result []data.Offering

	for _, v := range objects {
		result = append(result, *v.(*data.Offering))
	}

	return result, nil
}

// GetAgentOfferings returns active offerings available for a agent.
func (h *Handler) GetAgentOfferings(
	password, product, status string) ([]data.Offering, error) {
	logger := h.logger.Add("method", "GetAgentOfferings",
		"product", product, "status", status)

	if err := h.checkPassword(logger, password); err != nil {
		return nil, err
	}

	var (
		conditions string
		args       []interface{}
		index      = 1
	)

	if product != "" {
		conditions = fmt.Sprintf(
			"product = %s", h.db.Placeholder(index))
		args = append(args, product)
		index++
	}

	if status != "" {
		condition := fmt.Sprintf(
			"offer_status = %s", h.db.Placeholder(index))
		if conditions == "" {
			conditions = condition
		} else {
			conditions = fmt.Sprintf(
				"%s AND %s", conditions, condition)
		}
		args = append(args, status)
		index++
	}

	prefix := "WHERE "

	agentOfferingCondition := `
		agent IN (SELECT eth_addr FROM accounts)
			AND (SELECT in_use FROM accounts WHERE eth_addr = agent)
		      ORDER BY block_number_updated DESC`

	if conditions == "" {
		conditions = prefix + agentOfferingCondition
	} else {
		conditions = fmt.Sprintf("%s %s AND %s",
			prefix, conditions, agentOfferingCondition)
	}

	objects, err := h.selectAllFrom(
		logger, data.OfferingTable, conditions, args...)
	if err != nil {
		return nil, err
	}

	var result []data.Offering

	for _, v := range objects {
		result = append(result, *v.(*data.Offering))
	}

	return result, nil
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

	agentKey, err := h.decryptKeyFunc(agent.PrivateKey, h.pwdStorage.Get())
	if err != nil {
		return handleErr(err)
	}

	packed, err := messages.PackWithSignature(msgBytes, agentKey)
	if err != nil {
		return handleErr(err)
	}

	offering.RawMsg = data.FromBytes(packed)

	hashBytes := common.BytesToHash(crypto.Keccak256(packed))

	offering.Hash = data.FromBytes(hashBytes.Bytes())

	return nil
}

// fillOffering fills offerings nonce, status, hash and signature.
func (h *Handler) fillOffering(
	logger log.Logger, offering *data.Offering) error {
	agent := &data.Account{}
	if err := h.findByPrimaryKey(logger,
		ErrAccountNotFound, agent, offering.Agent); err != nil {
		return err
	}

	template := &data.Template{}
	if err := h.findByPrimaryKey(logger,
		ErrTemplateNotFound, template, offering.Template); err != nil {
		return err
	}

	offering.ID = util.NewUUID()
	offering.OfferStatus = data.OfferEmpty
	offering.Status = data.MsgUnpublished
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
func (h *Handler) UpdateOffering(password string,
	offering *data.Offering) error {
	logger := h.logger.Add(
		"method", "UpdateOffering", "offering", offering)

	err := h.checkPassword(logger, password)
	if err != nil {
		return err
	}

	err = h.findByPrimaryKey(logger, ErrOfferingNotFound, &data.Offering{}, offering.ID)
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
func (h *Handler) CreateOffering(password string,
	offering *data.Offering) (*string, error) {
	logger := h.logger.Add(
		"method", "CreateOffering", "offering", offering)

	err := h.checkPassword(logger, password)
	if err != nil {
		return nil, err
	}

	err = h.prepareOffering(logger, offering)
	if err != nil {
		return nil, err
	}

	err = insert(logger, h.db.Querier, offering)
	if err != nil {
		return nil, err
	}

	return &offering.ID, nil
}
