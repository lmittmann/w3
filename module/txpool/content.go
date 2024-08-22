package txpool

import (
	"encoding/json"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3/internal/module"
	"github.com/lmittmann/w3/w3types"
)

// Content requests the pending and queued transactions in the transaction pool.
func Content() w3types.RPCCallerFactory[*ContentResponse] {
	return module.NewFactory[*ContentResponse](
		"txpool_content",
		nil,
	)
}

// ContentFrom requests the pending and queued transactions in the transaction pool
// from the given address.
func ContentFrom(addr common.Address) w3types.RPCCallerFactory[*ContentFromResponse] {
	return module.NewFactory[*ContentFromResponse](
		"txpool_contentFrom",
		[]any{addr},
	)
}

type ContentResponse struct {
	Pending map[common.Address][]*types.Transaction
	Queued  map[common.Address][]*types.Transaction
}

func (c *ContentResponse) UnmarshalJSON(data []byte) error {
	type contentResponse struct {
		Pending map[common.Address]map[uint64]*types.Transaction `json:"pending"`
		Queued  map[common.Address]map[uint64]*types.Transaction `json:"queued"`
	}

	var dec contentResponse
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}

	c.Pending = make(map[common.Address][]*types.Transaction, len(dec.Pending))
	for addr, nonceTx := range dec.Pending {
		txs := make(types.TxByNonce, 0, len(nonceTx))
		for _, tx := range nonceTx {
			txs = append(txs, tx)
		}
		sort.Sort(txs)
		c.Pending[addr] = txs
	}

	c.Queued = make(map[common.Address][]*types.Transaction, len(dec.Queued))
	for addr, nonceTx := range dec.Queued {
		txs := make(types.TxByNonce, 0, len(nonceTx))
		for _, tx := range nonceTx {
			txs = append(txs, tx)
		}
		sort.Sort(txs)
		c.Queued[addr] = txs
	}

	return nil
}

type ContentFromResponse struct {
	Pending []*types.Transaction
	Queued  []*types.Transaction
}

func (cf *ContentFromResponse) UnmarshalJSON(data []byte) error {
	type contentFromResponse struct {
		Pending map[uint64]*types.Transaction `json:"pending"`
		Queued  map[uint64]*types.Transaction `json:"queued"`
	}

	var dec contentFromResponse
	if err := json.Unmarshal(data, &dec); err != nil {
		return err
	}

	txs := make(types.TxByNonce, 0, len(dec.Pending))
	for _, tx := range dec.Pending {
		txs = append(txs, tx)
	}
	sort.Sort(txs)
	cf.Pending = txs

	txs = make(types.TxByNonce, 0, len(dec.Queued))
	for _, tx := range dec.Queued {
		txs = append(txs, tx)
	}
	sort.Sort(txs)
	cf.Queued = txs

	return nil
}
