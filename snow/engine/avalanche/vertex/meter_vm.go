// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package vertex

import (
	"fmt"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/snow"
	"github.com/ava-labs/avalanchego/snow/consensus/snowstorm"
	"github.com/ava-labs/avalanchego/snow/engine/common"
	latencyMetrics "github.com/ava-labs/avalanchego/utils/metrics"
	"github.com/ava-labs/avalanchego/utils/timer"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	pending,
	parse,
	get prometheus.Histogram
}

func (m *metrics) Initialize(
	namespace string,
	registerer prometheus.Registerer,
) error {
	m.pending = latencyMetrics.NewNanosecnodsLatencyMetric(namespace, "pending_txs")
	m.parse = latencyMetrics.NewNanosecnodsLatencyMetric(namespace, "parse_tx")
	m.get = latencyMetrics.NewNanosecnodsLatencyMetric(namespace, "get_tx")

	errs := wrappers.Errs{}
	errs.Add(
		registerer.Register(m.pending),
		registerer.Register(m.parse),
		registerer.Register(m.get),
	)
	return errs.Err
}

type MeterVM struct {
	DAGVM
	metrics
	clock timer.Clock
}

func (vm *MeterVM) Initialize(
	ctx *snow.Context,
	db database.Database,
	genesisBytes []byte,
	toEngine chan<- common.Message,
	fxs []*common.Fx,
) error {
	if err := vm.metrics.Initialize(fmt.Sprintf("metervm_%s", ctx.Namespace), ctx.Metrics); err != nil {
		return err
	}

	return vm.DAGVM.Initialize(ctx, db, genesisBytes, toEngine, fxs)
}

func (vm *MeterVM) Pending() []snowstorm.Tx {
	start := vm.clock.Time()
	txs := vm.DAGVM.Pending()
	end := vm.clock.Time()
	vm.metrics.pending.Observe(float64(end.Sub(start)))
	return txs
}

func (vm *MeterVM) Parse(b []byte) (snowstorm.Tx, error) {
	start := vm.clock.Time()
	tx, err := vm.DAGVM.Parse(b)
	end := vm.clock.Time()
	vm.metrics.parse.Observe(float64(end.Sub(start)))
	return tx, err
}

func (vm *MeterVM) Get(txID ids.ID) (snowstorm.Tx, error) {
	start := vm.clock.Time()
	tx, err := vm.DAGVM.Get(txID)
	end := vm.clock.Time()
	vm.metrics.get.Observe(float64(end.Sub(start)))
	return tx, err
}
