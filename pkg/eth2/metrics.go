package eth2

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	ValidatorLabel = "validator"
)

var (
	metricsValidatorBalance = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "validator_balance",
			Help: "The Balance of validator",
		},
		[]string{ValidatorLabel},
	)
)
