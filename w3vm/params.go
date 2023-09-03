package w3vm

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

// allEthashProtocolChanges contains every protocol change introduced for Mainnet.
var allEthashProtocolChanges = &params.ChainConfig{
	ChainID:                       big.NewInt(1),
	HomesteadBlock:                new(big.Int),
	DAOForkBlock:                  nil,
	DAOForkSupport:                false,
	EIP150Block:                   new(big.Int),
	EIP155Block:                   new(big.Int),
	EIP158Block:                   new(big.Int),
	ByzantiumBlock:                new(big.Int),
	ConstantinopleBlock:           new(big.Int),
	PetersburgBlock:               new(big.Int),
	IstanbulBlock:                 new(big.Int),
	MuirGlacierBlock:              new(big.Int),
	BerlinBlock:                   new(big.Int),
	LondonBlock:                   new(big.Int),
	ArrowGlacierBlock:             new(big.Int),
	GrayGlacierBlock:              new(big.Int),
	MergeNetsplitBlock:            nil,
	ShanghaiTime:                  &uint0,
	CancunTime:                    nil,
	PragueTime:                    nil,
	VerkleTime:                    nil,
	TerminalTotalDifficulty:       nil,
	TerminalTotalDifficultyPassed: true,
	Ethash:                        new(params.EthashConfig),
	Clique:                        nil,
}

func defaultBlockContext() *vm.BlockContext {
	var coinbase common.Address
	rand.Read(coinbase[:])

	var random common.Hash
	rand.Read(random[:])

	return &vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		GetHash:     zeroHashFunc,
		Coinbase:    coinbase,
		BlockNumber: new(big.Int),
		Time:        uint64(time.Now().Unix()),
		Difficulty:  new(big.Int),
		BaseFee:     new(big.Int),
		GasLimit:    params.MaxGasLimit,
		Random:      &random,
	}
}
