package sesssrv

import (
	"encoding/json"
	"net/http"

	"github.com/privatix/dappctrl/util/log"
	"github.com/privatix/dappctrl/util/srv"
)

// Post posts a request with given arguments and returns a response result.
func Post(conf *srv.Config, logger log.Logger, username, password, path string,
	args, result interface{}) error {
	data, err := json.Marshal(args)
	if err != nil {
		logger.Error(err.Error())
		return ErrEncodeArgs
	}

	req, err := srv.NewHTTPRequest(
		conf, http.MethodPost, path, &srv.Request{Args: data})
	if err != nil {
		return err
	}

	req.SetBasicAuth(username, password)

	resp, err := srv.Send(req)
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return resp.Error
	}

	if resp.Result != nil && string(resp.Result) == "null" {
		return nil
	}

	if result != nil {
		err = json.Unmarshal(resp.Result, result)
		if err != nil {
			logger.Error(err.Error())
			return ErrDecodeResponse
		}
	}

	return nil
}
