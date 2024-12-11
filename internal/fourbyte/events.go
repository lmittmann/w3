// Code generated by "go generate"; DO NOT EDIT.
package fourbyte

import "github.com/lmittmann/w3"

var events = map[[32]byte]*w3.Event{
	{0x0d, 0x36, 0x48, 0xbd, 0x0f, 0x6b, 0xa8, 0x01, 0x34, 0xa3, 0x3b, 0xa9, 0x27, 0x5a, 0xc5, 0x85, 0xd9, 0xd3, 0x15, 0xf0, 0xad, 0x83, 0x55, 0xcd, 0xde, 0xfd, 0xe3, 0x1a, 0xfa, 0x28, 0xd0, 0xe9}: w3.MustNewEvent("PairCreated(address indexed token0, address indexed token1, address pair, uint256)"),
	{0x1c, 0x41, 0x1e, 0x9a, 0x96, 0xe0, 0x71, 0x24, 0x1c, 0x2f, 0x21, 0xf7, 0x72, 0x6b, 0x17, 0xae, 0x89, 0xe3, 0xca, 0xb4, 0xc7, 0x8b, 0xe5, 0x0e, 0x06, 0x2b, 0x03, 0xa9, 0xff, 0xfb, 0xba, 0xd1}: w3.MustNewEvent("Sync(uint112 reserve0, uint112 reserve1)"),
	{0x4c, 0x20, 0x9b, 0x5f, 0xc8, 0xad, 0x50, 0x75, 0x8f, 0x13, 0xe2, 0xe1, 0x08, 0x8b, 0xa5, 0x6a, 0x56, 0x0d, 0xff, 0x69, 0x0a, 0x1c, 0x6f, 0xef, 0x26, 0x39, 0x4f, 0x4c, 0x03, 0x82, 0x1c, 0x4f}: w3.MustNewEvent("Mint(address indexed sender, uint256 amount0, uint256 amount1)"),
	{0x8c, 0x5b, 0xe1, 0xe5, 0xeb, 0xec, 0x7d, 0x5b, 0xd1, 0x4f, 0x71, 0x42, 0x7d, 0x1e, 0x84, 0xf3, 0xdd, 0x03, 0x14, 0xc0, 0xf7, 0xb2, 0x29, 0x1e, 0x5b, 0x20, 0x0a, 0xc8, 0xc7, 0xc3, 0xb9, 0x25}: w3.MustNewEvent("Approval(address indexed owner, address indexed spender, uint256 value)"),
	{0xd7, 0x8a, 0xd9, 0x5f, 0xa4, 0x6c, 0x99, 0x4b, 0x65, 0x51, 0xd0, 0xda, 0x85, 0xfc, 0x27, 0x5f, 0xe6, 0x13, 0xce, 0x37, 0x65, 0x7f, 0xb8, 0xd5, 0xe3, 0xd1, 0x30, 0x84, 0x01, 0x59, 0xd8, 0x22}: w3.MustNewEvent("Swap(address indexed sender, uint256 amount0In, uint256 amount1In, uint256 amount0Out, uint256 amount1Out, address indexed to)"),
	{0xdc, 0xcd, 0x41, 0x2f, 0x0b, 0x12, 0x52, 0x81, 0x9c, 0xb1, 0xfd, 0x33, 0x0b, 0x93, 0x22, 0x4c, 0xa4, 0x26, 0x12, 0x89, 0x2b, 0xb3, 0xf4, 0xf7, 0x89, 0x97, 0x6e, 0x6d, 0x81, 0x93, 0x64, 0x96}: w3.MustNewEvent("Burn(address indexed sender, uint256 amount0, uint256 amount1, address indexed to)"),
	{0xdd, 0xf2, 0x52, 0xad, 0x1b, 0xe2, 0xc8, 0x9b, 0x69, 0xc2, 0xb0, 0x68, 0xfc, 0x37, 0x8d, 0xaa, 0x95, 0x2b, 0xa7, 0xf1, 0x63, 0xc4, 0xa1, 0x16, 0x28, 0xf5, 0x5a, 0x4d, 0xf5, 0x23, 0xb3, 0xef}: w3.MustNewEvent("Transfer(address indexed from, address indexed to, uint256 value)"),
}