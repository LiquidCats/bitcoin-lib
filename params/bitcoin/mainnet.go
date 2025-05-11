package bitcoin

import "github.com/LiquidCats/bitcoin-lib/params"

const MainNet params.BitcoinNet = 0xd9b4bef9

var MainNetParams = params.Params{
	Name:        "mainnet",
	Net:         MainNet,
	DefaultPort: "8333",
	DNSSeeds: []params.DNSSeed{
		{"seed.bitcoin.sipa.be", true},
		{"dnsseed.bluematt.me", true},
		{"dnsseed.bitcoin.dashjr.org", false},
		{"seed.bitnodes.io", false},
		{"seed.bitcoin.jonasschnelli.ch", true},
		{"seed.btc.petertodd.net", true},
		{"seed.bitcoin.sprovoost.nl", true},
		{"seed.bitcoin.wiz.biz", true},
	},

	// Mempool parameters
	RelayNonStdTxs: false,

	// Human-readable part for Bech32 encoded segwit addresses, as defined in
	// BIP 173.
	Bech32HRPSegwit: "bc", // always bc for main net

	// Address encoding magics
	PubKeyHashAddrID:        0x00, // starts with 1
	ScriptHashAddrID:        0x05, // starts with 3
	PrivateKeyID:            0x80, // starts with 5 (uncompressed) or K (compressed)
	WitnessPubKeyHashAddrID: 0x06, // starts with p2
	WitnessScriptHashAddrID: 0x0A, // starts with 7Xh

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x04, 0x88, 0xad, 0xe4}, // starts with xprv
	HDPublicKeyID:  [4]byte{0x04, 0x88, 0xb2, 0x1e}, // starts with xpub

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	HDCoinType: 0,
}
