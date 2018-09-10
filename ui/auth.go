package ui

import (
	"github.com/privatix/dappctrl/data"
	"github.com/privatix/dappctrl/util"
	"github.com/privatix/dappctrl/util/log"
)

// SetPassword sets the password only on the first call.
// Returns error if password already exists.
func (h *Handler) SetPassword(password string) error {
	logger := h.logger.Add("method", "SetPassword")

	if password == "" {
		logger.Warn("received empty password")
		return ErrEmptyPassword
	}

	if err := h.validatePasswordNotSet(logger); err != nil {
		return err
	}

	salt := util.NewUUID()

	hashed, err := hashedPassword(logger, password, salt)
	if err != nil {
		return err
	}

	tx, err := beginTX(logger, h.db)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := insert(logger, tx.Querier,
		saltSetting(salt), passwordHashSetting(hashed)); err != nil {
		return err
	}

	return commitTX(logger, tx)
}

// UpdatePassword updates the password.
func (h *Handler) UpdatePassword(current, new string) error {
	logger := h.logger.Add("method", "UpdatePassword")

	if err := h.checkPassword(logger, current); err != nil {
		return err
	}

	if new == "" {
		logger.Warn("received empty password for update")
		return ErrEmptyPassword
	}

	salt := util.NewUUID()

	hashed, err := hashedPassword(logger, new, salt)
	if err != nil {
		return err
	}

	tx, err := beginTX(logger, h.db)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := update(logger, tx.Querier,
		saltSetting(salt), passwordHashSetting(hashed)); err != nil {
		return err
	}

	return commitTX(logger, tx)
}

func (h *Handler) validatePasswordNotSet(logger log.Logger) error {
	var count int
	err := h.db.QueryRow("SELECT count(*) FROM settings WHERE key in ($1, $2)",
		data.SettingPasswordSalt, data.SettingPasswordHash).Scan(&count)
	if err != nil {
		logger.Error(err.Error())
		return ErrInternal
	}
	if count > 0 {
		logger.Warn("received repeated set password")
		return ErrPasswordExists
	}
	return nil
}

func hashedPassword(logger log.Logger, password, salt string) (string, error) {
	hashed, err := data.HashPassword(password, salt)
	if err != nil {
		logger.Error(err.Error())
		return "", ErrInternal
	}
	return hashed, nil
}

func saltSetting(salt string) *data.Setting {
	return &data.Setting{
		Key:         data.SettingPasswordSalt,
		Value:       salt,
		Permissions: data.AccessDenied,
		Name:        "Salt",
	}
}

func passwordHashSetting(hash string) *data.Setting {
	return &data.Setting{
		Key:         data.SettingPasswordHash,
		Value:       hash,
		Permissions: data.AccessDenied,
		Name:        "Password",
	}
}
