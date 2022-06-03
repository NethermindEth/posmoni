package utils

import (
	"io"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v4"
	log "github.com/sirupsen/logrus"
)

/*
GetRequest :
Make a GET request to the given URL. Uses exponential retries with backoff.

params :-
a. url string
URL to make the request to
b. retryDuration time.Duration
Duration to wait between retries

returns :-
a. http.Response
Response from the request
b. error
Error if any
*/
func GetRequest(url string, retryDuration time.Duration) (*http.Response, error) {
	logFields := log.Fields{"Method": "GetRequest"}
	var response *http.Response

	// Adding exponential retry
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = retryDuration

	err := backoff.Retry(func() (err error) {
		response, err = http.Get(url)
		if err != nil {
			log.WithFields(logFields).Errorf("request failed. Error: %v", err)
			log.WithFields(logFields).Info("Retrying request")
			return err
		} else if response.StatusCode != 200 {
			log.WithFields(logFields).Errorf("bad response, got: %d", response.StatusCode)
		}
		return nil
	}, b)

	if err != nil {
		return nil, err
	}

	return response, nil
}

/*
PostRequest :
Make a POST request to the given URL. Uses exponential retries with backoff optionally.

params :-
a. url string
URL to make the request to
b. retry bool
True if retries should be done
c. retryDuration time.Duration
Duration to wait between retries

returns :-
a. http.Response
Response from the request
b. error
Error if any
*/
func PostRequest(url, contentType string, body io.Reader, retry bool, retryDuration time.Duration) (*http.Response, error) {
	logFields := log.Fields{"Method": "PostRequest"}
	var response *http.Response
	var err error

	if retry {
		// Adding exponential retry
		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = retryDuration

		err = backoff.Retry(func() (err error) {
			response, err = http.Post(url, contentType, body)
			if err != nil {
				log.WithFields(logFields).Errorf("request failed. Error: %v", err)
				log.WithFields(logFields).Info("Retrying request")
				return err
			} else if response.StatusCode != 200 {
				log.WithFields(logFields).Errorf("bad response, got: %d", response.StatusCode)
			}
			return nil
		}, b)
	} else {
		response, err = http.Post(url, contentType, body)
	}

	if err != nil {
		return nil, err
	}
	return response, nil
}
