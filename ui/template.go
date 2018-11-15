package ui

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/reform.v1"

	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

// GetTemplates returns templates.
func (h *Handler) GetTemplates(
	password, tplType string) ([]data.Template, error) {
	logger := h.logger.Add(
		"method", "GetTemplates", "type", tplType)

	err := h.checkPassword(logger, password)
	if err != nil {
		return nil, err
	}

	var templates []reform.Struct

	if tplType != "" {
		templates, err = h.selectAllFrom(
			logger, data.TemplateTable,
			"WHERE kind = $1", tplType)
	} else {
		templates, err = h.selectAllFrom(
			logger, data.TemplateTable, "")
	}

	if err != nil {
		return nil, err
	}

	result := make([]data.Template, 0)

	for _, v := range templates {
		result = append(result, *v.(*data.Template))
	}

	return result, nil
}

func checkTemplate(logger log.Logger, template *data.Template) error {
	v := make(map[string]interface{})
	if template.Kind != data.TemplateOffer &&
		template.Kind != data.TemplateAccess {
		logger.Error("invalid template type")
		return ErrInvalidTemplateType
	}
	if err := json.Unmarshal(template.Raw, &v); err != nil {
		logger.Error(err.Error())
		return ErrMalformedTemplate
	}
	return nil
}

// CreateTemplate creates template.
func (h *Handler) CreateTemplate(
	password string, template *data.Template) (*string, error) {
	logger := h.logger.Add("method", "CreateTemplate",
		"template", template)

	err := h.checkPassword(logger, password)
	if err != nil {
		return nil, err
	}

	err = checkTemplate(logger, template)
	if err != nil {
		return nil, err
	}

	template.ID = util.NewUUID()
	template.Hash = data.HexFromBytes(crypto.Keccak256(template.Raw))

	err = h.insertObject(template)
	if err != nil {
		return nil, err
	}
	return &template.ID, nil
}
