package utils

import (
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	log "github.com/sirupsen/logrus"
)

func GetRequest(url string, retryDuration time.Duration) (*http.Response, error) {
	var response *http.Response

	// Adding exponential retry
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = retryDuration

	err := backoff.Retry(func() (err error) {
		response, err = http.Get(url)
		if err != nil {
			log.Errorf("request failed. Error: %v", err)
			log.Info("Retrying request")
			return err
		} else if response.StatusCode != 200 {
			log.Errorf("bad response, got: %d", response.StatusCode)
		}
		return nil
	}, b)

	if err != nil {
		return nil, err
	}

	return response, nil
}
