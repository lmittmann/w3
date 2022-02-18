package eth

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

//go:generate gencodec -type RPCTransaction -field-override rpcTransactionMarshaling -out types_transaction.gen.go
type RPCTransaction struct {
	BlockHash        *common.Hash      `json:"blockHash"`
	BlockNumber      *big.Int          `json:"blockNumber"`
	From             common.Address    `json:"from"`
	Gas              uint64            `json:"gas"`
	GasPrice         *big.Int          `json:"gasPrice"`
	GasFeeCap        *big.Int          `json:"maxFeePerGas,omitempty"`
	GasTipCap        *big.Int          `json:"maxPriorityFeePerGas,omitempty"`
	Hash             common.Hash       `json:"hash"`
	Input            []byte            `json:"input"`
	Nonce            uint64            `json:"nonce"`
	To               *common.Address   `json:"to"`
	TransactionIndex *uint64           `json:"transactionIndex"`
	Value            *big.Int          `json:"value"`
	Type             uint64            `json:"type"`
	Accesses         *types.AccessList `json:"accessList,omitempty"`
	ChainID          *big.Int          `json:"chainId,omitempty"`
	V                *big.Int          `json:"v"`
	R                *big.Int          `json:"r"`
	S                *big.Int          `json:"s"`
}

type rpcTransactionMarshaling struct {
	BlockHash        *common.Hash
	BlockNumber      *hexutil.Big
	From             common.Address
	Gas              hexutil.Uint64
	GasPrice         *hexutil.Big
	GasFeeCap        *hexutil.Big
	GasTipCap        *hexutil.Big
	Hash             common.Hash
	Input            hexutil.Bytes
	Nonce            hexutil.Uint64
	To               *common.Address
	TransactionIndex *hexutil.Uint64
	Value            *hexutil.Big
	Type             hexutil.Uint64
	Accesses         *types.AccessList
	ChainID          *hexutil.Big
	V                *hexutil.Big
	R                *hexutil.Big
	S                *hexutil.Big
}

//go:generate gencodec -type RPCReceipt -field-override rpcReceiptMarshaling -out types_receipt.gen.go
type RPCReceipt struct {
	TransactionHash   common.Hash     `json:"transactionHash"`
	TransactionIndex  uint64          `json:"transactionIndex"`
	BlockHash         common.Hash     `json:"blockHash"`
	BlockNumber       *big.Int        `json:"blockNumber"`
	From              common.Address  `json:"from"`
	To                *common.Address `json:"to"`
	CumulativeGasUsed uint64          `json:"cumulativeGasUsed"`
	GasUsed           uint64          `json:"gasUsed"`
	ContractAddress   *common.Address `json:"contractAddress"`
	Logs              []*types.Log    `json:"logs"`
	LogsBloom         types.Bloom     `json:"logsBloom"`
	Type              uint64          `json:"type"`
	Status            uint64          `json:"status"` // 1 (success) or 0 (failure)
}

type rpcReceiptMarshaling struct {
	TransactionHash   common.Hash
	TransactionIndex  hexutil.Uint64
	BlockHash         common.Hash
	BlockNumber       *hexutil.Big
	From              common.Address
	To                *common.Address
	CumulativeGasUsed hexutil.Uint64
	GasUsed           hexutil.Uint64
	ContractAddress   *common.Address
	Logs              []*types.Log
	LogsBloom         types.Bloom
	Type              hexutil.Uint64
	Status            hexutil.Uint64
}

//go:generate gencodec -type RPCHeader -field-override rpcHeaderMarshaling -out types_header.gen.go
type RPCHeader struct {
	Hash              common.Hash   `json:"hash"`
	TransactionHashes []common.Hash `json:"transactions"`
	UncleHashes       []common.Hash `json:"uncles"`

	ParentHash  common.Hash      `json:"parentHash"       gencodec:"required"`
	UncleHash   common.Hash      `json:"sha3Uncles"       gencodec:"required"`
	Coinbase    common.Address   `json:"miner"            gencodec:"required"`
	Root        common.Hash      `json:"stateRoot"        gencodec:"required"`
	TxHash      common.Hash      `json:"transactionsRoot" gencodec:"required"`
	ReceiptHash common.Hash      `json:"receiptsRoot"     gencodec:"required"`
	Bloom       types.Bloom      `json:"logsBloom"        gencodec:"required"`
	Difficulty  *big.Int         `json:"difficulty"       gencodec:"required"`
	Number      *big.Int         `json:"number"           gencodec:"required"`
	GasLimit    uint64           `json:"gasLimit"         gencodec:"required"`
	GasUsed     uint64           `json:"gasUsed"          gencodec:"required"`
	Time        uint64           `json:"timestamp"        gencodec:"required"`
	Extra       []byte           `json:"extraData"        gencodec:"required"`
	MixDigest   common.Hash      `json:"mixHash"`
	Nonce       types.BlockNonce `json:"nonce"`
	BaseFee     *big.Int         `json:"baseFeePerGas" rlp:"optional"`
}

type rpcHeaderMarshaling struct {
	Hash              common.Hash
	TransactionHashes []common.Hash
	UncleHashes       []common.Hash

	ParentHash  common.Hash
	UncleHash   common.Hash
	Coinbase    common.Address
	Root        common.Hash
	TxHash      common.Hash
	ReceiptHash common.Hash
	Bloom       types.Bloom
	Difficulty  *hexutil.Big
	Number      *hexutil.Big
	GasLimit    hexutil.Uint64
	GasUsed     hexutil.Uint64
	Time        hexutil.Uint64
	Extra       hexutil.Bytes
	MixDigest   common.Hash
	Nonce       types.BlockNonce
	BaseFee     *hexutil.Big
}

//go:generate gencodec -type RPCBlock -field-override rpcBlockMarshaling -out types_block.gen.go
type RPCBlock struct {
	Hash         common.Hash      `json:"hash"`
	Transactions []RPCTransaction `json:"transactions"`
	UncleHashes  []common.Hash    `json:"uncles"`

	ParentHash  common.Hash      `json:"parentHash"       gencodec:"required"`
	UncleHash   common.Hash      `json:"sha3Uncles"       gencodec:"required"`
	Coinbase    common.Address   `json:"miner"            gencodec:"required"`
	Root        common.Hash      `json:"stateRoot"        gencodec:"required"`
	TxHash      common.Hash      `json:"transactionsRoot" gencodec:"required"`
	ReceiptHash common.Hash      `json:"receiptsRoot"     gencodec:"required"`
	Bloom       types.Bloom      `json:"logsBloom"        gencodec:"required"`
	Difficulty  *big.Int         `json:"difficulty"       gencodec:"required"`
	Number      *big.Int         `json:"number"           gencodec:"required"`
	GasLimit    uint64           `json:"gasLimit"         gencodec:"required"`
	GasUsed     uint64           `json:"gasUsed"          gencodec:"required"`
	Time        uint64           `json:"timestamp"        gencodec:"required"`
	Extra       []byte           `json:"extraData"        gencodec:"required"`
	MixDigest   common.Hash      `json:"mixHash"`
	Nonce       types.BlockNonce `json:"nonce"`
	BaseFee     *big.Int         `json:"baseFeePerGas" rlp:"optional"`
}

type rpcBlockMarshaling struct {
	Hash         common.Hash
	Transactions []RPCTransaction
	UncleHashes  []common.Hash

	ParentHash  common.Hash
	UncleHash   common.Hash
	Coinbase    common.Address
	Root        common.Hash
	TxHash      common.Hash
	ReceiptHash common.Hash
	Bloom       types.Bloom
	Difficulty  *hexutil.Big
	Number      *hexutil.Big
	GasLimit    hexutil.Uint64
	GasUsed     hexutil.Uint64
	Time        hexutil.Uint64
	Extra       hexutil.Bytes
	MixDigest   common.Hash
	Nonce       types.BlockNonce
	BaseFee     *hexutil.Big
}
