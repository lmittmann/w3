package eth_test

import (
	"fmt"
	"math/big"
	"sync/atomic"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/internal/rpctest"
	"github.com/lmittmann/w3/module/eth"
)

func TestBlockByNumber__1(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/get_block_by_number__1.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		block     = new(types.Block)
		wantBlock = types.NewBlockWithHeader(&types.Header{
			ParentHash:  w3.H("0xd4e56740f876aef8c010b86a40d5f56745a118d0906a34e69aec8c0db1cb8fa3"),
			UncleHash:   w3.H("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    w3.A("0x05a56E2D52c817161883f50c441c3228CFe54d9f"),
			Root:        w3.H("0xd67e4d450343046425ae4271474353857ab860dbc0a1dde64b41b5cd3a532bf3"),
			TxHash:      w3.H("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"),
			ReceiptHash: w3.H("0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"),
			Bloom:       types.BytesToBloom(w3.B("0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
			Difficulty:  w3.I("0x3ff800000"),
			Number:      w3.I("0x1"),
			GasLimit:    0x1388,
			GasUsed:     0x0,
			Time:        0x55ba4224,
			Extra:       w3.B("0x476574682f76312e302e302f6c696e75782f676f312e342e32"),
			MixDigest:   w3.H("0x969b900de27b6ac6a67742365dd65f55a0526c41fd18e1b16f1a1215c2e66f59"),
			Nonce:       types.EncodeNonce(0x539bd4979fef1ec4),
		})
	)

	if err := client.Call(eth.BlockByNumber(big.NewInt(1)).Returns(block)); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if diff := cmp.Diff(wantBlock, block,
		cmp.AllowUnexported(big.Int{}, types.Block{}, atomic.Value{}),
		cmpopts.EquateEmpty()); diff != "" {
		t.Fatalf("(-want, +got)\n%s", diff)
	}
}

func TestBlockByNumber__46147(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/get_block_by_number__46147.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		block     = new(types.Block)
		wantBlock = types.NewBlockWithHeader(&types.Header{
			ParentHash:  w3.H("0x5a41d0e66b4120775176c09fcf39e7c0520517a13d2b57b18d33d342df038bfc"),
			UncleHash:   w3.H("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    w3.A("0xe6A7a1d47ff21B6321162AEA7C6CB457D5476Bca"),
			Root:        w3.H("0x0e0df2706b0a4fb8bd08c9246d472abbe850af446405d9eba1db41db18b4a169"),
			TxHash:      w3.H("0x4513310fcb9f6f616972a3b948dc5d547f280849a87ebb5af0191f98b87be598"),
			ReceiptHash: w3.H("0xfe2bf2a941abf41d72637e5b91750332a30283efd40c424dc522b77e6f0ed8c4"),
			Bloom:       types.BytesToBloom(w3.B("0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")),
			Difficulty:  w3.I("0x153886c1bbd"),
			Number:      w3.I("0xb443"),
			GasLimit:    0x520b,
			GasUsed:     0x5208,
			Time:        0x55c42659,
			Extra:       w3.B("0x657468706f6f6c2e6f7267"),
			MixDigest:   w3.H("0xb48c515a9dde8d346c3337ea520aa995a4738bb595495506125449c1149d6cf4"),
			Nonce:       types.EncodeNonce(0xba4f8ecd18aab215),
		}).WithBody(
			types.Transactions{
				types.NewTx(&types.LegacyTx{
					Nonce:    0x0,
					GasPrice: w3.I("0x2d79883d2000"),
					Gas:      0x5208,
					To:       w3.APtr("0x5DF9B87991262F6BA471F09758CDE1c0FC1De734"),
					Value:    w3.I("0x7a69"),
					V:        w3.I("0x1c"),
					R:        w3.I("0x88ff6cf0fefd94db46111149ae4bfc179e9b94721fffd821d38d16464b3f71d0"),
					S:        w3.I("0x45e0aff800961cfce805daef7016b9b675c137a6a41a548f7b60a3484c06a33a"),
				}),
			}, nil,
		)
	)

	if err := client.Call(eth.BlockByNumber(big.NewInt(46147)).Returns(block)); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if diff := cmp.Diff(wantBlock, block,
		cmp.AllowUnexported(big.Int{}, types.Block{}, types.Transaction{}, atomic.Value{}),
		cmpopts.IgnoreFields(types.Transaction{}, "time"),
		cmpopts.EquateEmpty()); diff != "" {
		t.Fatalf("(-want, +got)\n%s", diff)
	}
}

func TestHeaderByNumber__12965000(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/get_block_by_number__12965000.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		header     = new(types.Header)
		wantHeader = &types.Header{
			ParentHash:  w3.H("0x3de6bb3849a138e6ab0b83a3a00dc7433f1e83f7fd488e4bba78f2fe2631a633"),
			UncleHash:   w3.H("0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"),
			Coinbase:    w3.A("0x7777788200b672a42421017f65ede4fc759564c8"),
			Root:        w3.H("0x41cf6e8e60fd087d2b00360dc29e5bfb21959bce1f4c242fd1ad7c4da968eb87"),
			TxHash:      w3.H("0xdfcb68d3a3c41096f4a77569db7956e0a0e750fad185948e54789ea0e51779cb"),
			ReceiptHash: w3.H("0x8a8865cd785e2e9dfce7da83aca010b10b9af2abbd367114b236f149534c821d"),
			Bloom:       types.BytesToBloom(w3.B("0x24e74ad77d9a2b27bdb8f6d6f7f1cffdd8cfb47fdebd433f011f7dfcfbb7db638fadd5ff66ed134ede2879ce61149797fbcdf7b74f6b7de153ec61bdaffeeb7b59c3ed771a2fe9eaed8ac70e335e63ff2bfe239eaff8f94ca642fdf7ee5537965be99a440f53d2ce057dbf9932be9a7b9a82ffdffe4eeee1a66c4cfb99fe4540fbff936f97dde9f6bfd9f8cefda2fc174d23dfdb7d6f7dfef5f754fe6a7eec92efdbff779b5feff3beafebd7fd6e973afebe4f5d86f3aafb1f73bf1e1d0cdd796d89827edeffe8fb6ae6d7bf639ec5f5ff4c32f31f6b525b676c7cdf5e5c75bfd5b7bd1928b6f43aac7fa0f6336576e5f7b7dfb9e8ebbe6f6efe2f9dfe8b3f56")),
			Difficulty:  w3.I("0x1b81c1fe05b218"),
			Number:      w3.I("0xc5d488"),
			GasLimit:    0x1ca3542,
			GasUsed:     0x1ca2629,
			Time:        0x610bdaa6,
			Extra:       w3.B("0x68747470733a2f2f7777772e6b7279707465782e6f7267"),
			MixDigest:   w3.H("0x9620b46a81a4795cf4449d48e3270419f58b09293a5421205f88179b563f815a"),
			Nonce:       types.EncodeNonce(0xb223da049adf2216),
			BaseFee:     w3.I("0x3b9aca00"),
		}
	)

	if err := client.Call(eth.HeaderByNumber(big.NewInt(12965000)).Returns(header)); err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	if diff := cmp.Diff(wantHeader, header,
		cmp.AllowUnexported(big.Int{})); diff != "" {
		t.Fatalf("(-want, +got)\n%s", diff)
	}
}

func TestBlockByNumber__999999999(t *testing.T) {
	t.Parallel()

	srv := rpctest.NewFileServer(t, "testdata/get_block_by_number__999999999.golden")
	defer srv.Close()

	client := w3.MustDial(srv.URL())
	defer client.Close()

	var (
		block   = new(types.Block)
		wantErr = fmt.Errorf("w3: response handling failed: not found")
	)

	if gotErr := client.Call(eth.BlockByNumber(big.NewInt(999999999)).Returns(block)); wantErr.Error() != gotErr.Error() {
		t.Fatalf("want %v, got %v", wantErr, gotErr)
	}
}
