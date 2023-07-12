package w3vm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

var (
	uint0 uint64
)

// defaultChainConfig contains every protocol change introduced for Mainnet.
var defaultChainConfig = &params.ChainConfig{
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
