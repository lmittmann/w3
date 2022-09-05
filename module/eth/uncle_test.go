package eth_test

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

var header15050036 = types.Header{
	BaseFee:     big.NewInt(0xa688e9452),
	Difficulty:  big.NewInt(0x3252ebd268b855),
	Extra:       w3.B("0x466c6578706f6f6c2f53302f484b202d20546f726f6e746f"),
	GasLimit:    0x1c9c380,
	GasUsed:     0x47d8e7,
	Bloom:       blockBloom(w3.B("0x08200144090018000030b040821105200400800000148002000106861402e08111400008080b0420124248000008c1441a1110002a00361080c00010053a202124420000000020286900003ff080402082e6000804459404b28029019201102a1300f000020ac02082820580182028402115c420e030040288000030082a40129210210500030010005c0000020020040c10408529221058c0002040401028059a800123802664080ac0428105350210010402004010004330280204000b8840111500020400003828080000a000604108004260000900107400192a000920048114354818318008000016804104181020a210280068044000480a01020a6083")),
	Coinbase:    w3.A("0x7F101fE45e6649A6fB8F3F8B43ed03D353f2B90c"),
	MixDigest:   w3.H("0x99cabdd1355726d424faa49bf83b1820d456bb0f7d5fe83d53cff777294b4fe4"),
	Nonce:       blockNonce(w3.B("0xfbe8b3c8c41a4459")),
	Root:        w3.H("0x4143920b2a110d8956ab2edd3f6be5042c6c25a62e79d94863bd54e0e54f38da"),
	ReceiptHash: w3.H("0x8bf939b3ced21c57210ef6935ed9add6803c696adab0c2be1c54e2239c674ffe"),
	Number:      big.NewInt(0xe5a533),
	ParentHash:  w3.H("0x8fb655702f8f86f047e9acc49842efee36f96a1ddf620d0f4ee312723f4908f7"),
	TxHash:      w3.H("0xb57bedbc9fa725a7a765bc833ee3b1e946816960780693415e144bf78a9f0a95"),
	UncleHash:   w3.H("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
	Time:        0x62bd8341,
}

func blockNonce(data []byte) (nonce types.BlockNonce) {
	copy(nonce[:], data[:])
	return
}

func blockBloom(data []byte) (bloom types.Bloom) {
	copy(bloom[:], data[:])
	return
}

func TestUncleByBlockHashAndIndex(t *testing.T) {
	tests := []testCase[types.Header]{
		{
			Golden:  "uncle_by_hash_and_index__15050036",
			Call:    eth.UncleByBlockHashAndIndex(w3.H("0x7a98a492c1288a8451905bc665cb28d45fbdf8913c34d4ad756acb0609342e67"), 0),
			WantRet: header15050036,
		},
		{
			Golden:  "uncle_by_hash_and_index__15050036_1",
			Call:    eth.UncleByBlockHashAndIndex(w3.H("0x7a98a492c1288a8451905bc665cb28d45fbdf8913c34d4ad756acb0609342e67"), 1),
			WantErr: errors.New("w3: response handling failed: not found"),
		},
	}

	runTestCases(t, tests)
}

func TestUncleByBlockNumberAndIndex(t *testing.T) {
	tests := []testCase[types.Header]{
		{
			Golden:  "uncle_by_number_and_index__15050036",
			Call:    eth.UncleByBlockNumberAndIndex(big.NewInt(15050036), 0),
			WantRet: header15050036,
		},
	}

	runTestCases(t, tests)
}

func TestUncleCountByBlockHash(t *testing.T) {
	tests := []testCase[uint]{
		{
			Golden:  "uncle_count_by_hash__15050036",
			Call:    eth.UncleCountByBlockHash(w3.H("0x7a98a492c1288a8451905bc665cb28d45fbdf8913c34d4ad756acb0609342e67")),
			WantRet: 1,
		},
	}

	runTestCases(t, tests)
}

func TestUncleCountByBlockNumber(t *testing.T) {
	tests := []testCase[uint]{
		{
			Golden:  "uncle_count_by_number__15050036",
			Call:    eth.UncleCountByBlockNumber(big.NewInt(15050036)),
			WantRet: 1,
		},
	}

	runTestCases(t, tests)
}
