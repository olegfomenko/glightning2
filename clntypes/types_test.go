package clntypes

import (
	"encoding/json"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func TestHashJSON(t *testing.T) {
	const raw = `"000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"`

	var got Hash
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("unmarshal hash: %v", err)
	}
	if got[0] != 0x00 || got[31] != 0x1f {
		t.Fatalf("unexpected hash bytes: %x", got)
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal hash: %v", err)
	}
	if string(encoded) != raw {
		t.Fatalf("marshal hash = %s, want %s", encoded, raw)
	}
}

func TestSecretJSON(t *testing.T) {
	const raw = `"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"`

	var got Secret
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("unmarshal secret: %v", err)
	}
	if got[0] != 0xff || got[31] != 0xff {
		t.Fatalf("unexpected secret bytes: %x", got)
	}
}

func TestTxIDMappingUsesChainhash(t *testing.T) {
	const raw = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

	var got TxID
	if err := got.UnmarshalJSON([]byte(`"` + raw + `"`)); err != nil {
		t.Fatalf("unmarshal txid: %v", err)
	}
	if got.String() != raw {
		t.Fatalf("txid string = %s, want %s", got.String(), raw)
	}

	var _ chainhash.Hash = got
}

func TestOutpointMappingUsesWireOutPoint(t *testing.T) {
	const raw = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f:7"

	got, err := wire.NewOutPointFromString(raw)
	if err != nil {
		t.Fatalf("parse outpoint: %v", err)
	}
	if got.String() != raw {
		t.Fatalf("outpoint string = %s, want %s", got.String(), raw)
	}
}

func TestShortChannelID(t *testing.T) {
	got, err := ParseShortChannelID("123x456x7")
	if err != nil {
		t.Fatalf("parse scid: %v", err)
	}
	if got.BlockHeight() != 123 || got.TxIndex() != 456 || got.OutputIndex() != 7 {
		t.Fatalf("unexpected scid components: %d %d %d", got.BlockHeight(), got.TxIndex(), got.OutputIndex())
	}
	if got.String() != "123x456x7" {
		t.Fatalf("scid string = %s", got.String())
	}

	var fromJSON ShortChannelID
	if err := json.Unmarshal([]byte(`"123x456x7"`), &fromJSON); err != nil {
		t.Fatalf("unmarshal scid: %v", err)
	}
	if fromJSON != got {
		t.Fatalf("unmarshal scid = %v, want %v", fromJSON, got)
	}
}

func TestShortChannelIDDir(t *testing.T) {
	got, err := ParseShortChannelIDDir("123x456x7/1")
	if err != nil {
		t.Fatalf("parse scid dir: %v", err)
	}
	if got.Direction != 1 || got.ShortChannelID.String() != "123x456x7" {
		t.Fatalf("unexpected scid dir: %+v", got)
	}
	if got.String() != "123x456x7/1" {
		t.Fatalf("scid dir string = %s", got.String())
	}
}

func TestAmountOrAll(t *testing.T) {
	tests := map[string]uint64{
		`42`:          42,
		`"42sat"`:     42,
		`"42000msat"`: 42,
		`"0.01btc"`:   1000000,
	}
	for input, want := range tests {
		var got AmountOrAll
		if err := json.Unmarshal([]byte(input), &got); err != nil {
			t.Fatalf("unmarshal amount %s: %v", input, err)
		}
		if got.All || got.Amount.Uint64() != want {
			t.Fatalf("amount %s = %+v, want %d", input, got, want)
		}
	}

	var all AmountOrAll
	if err := json.Unmarshal([]byte(`"all"`), &all); err != nil {
		t.Fatalf("unmarshal all: %v", err)
	}
	if !all.All {
		t.Fatalf("expected all literal")
	}
}

func TestAmountOrAny(t *testing.T) {
	var got AmountOrAny
	if err := json.Unmarshal([]byte(`"any"`), &got); err != nil {
		t.Fatalf("unmarshal any: %v", err)
	}
	if !got.Any {
		t.Fatalf("expected any literal")
	}
}

func TestOutputDesc(t *testing.T) {
	var got OutputDesc
	if err := json.Unmarshal([]byte(`{"bc1qexample":"1000sat","bc1qall":"all"}`), &got); err != nil {
		t.Fatalf("unmarshal outputdesc: %v", err)
	}
	if got["bc1qexample"].Amount.Uint64() != 1000 {
		t.Fatalf("unexpected output amount: %+v", got["bc1qexample"])
	}
	if !got["bc1qall"].All {
		t.Fatalf("expected all output")
	}
}

func TestFeerateIsString(t *testing.T) {
	var got Feerate = "253perkw"
	if got != "253perkw" {
		t.Fatalf("feerate = %q", got)
	}
}
