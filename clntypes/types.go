package clntypes

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type (
	// Hex is an arbitrary hex-encoded byte string. CLN schemas use this when
	// the field length is defined by the specific RPC field rather than by the
	// scalar type itself.
	Hex = string

	// Currency is a decimal amount with a three-letter uppercase currency code,
	// such as "12.34USD".
	Currency = string

	// PubKey is a compressed secp256k1 public key, used by CLN for node IDs and
	// other Lightning public keys. Its JSON form is 33 bytes of SEC1 compressed
	// public key data encoded as hex.
	PubKey btcec.PublicKey

	// Signature is a secp256k1 ECDSA signature. CLN's JSON schema represents it
	// as DER-encoded signature bytes encoded as hex.
	Signature ecdsa.Signature

	// Bip340Sig is a BIP340 Schnorr signature. CLN's JSON form is the 64-byte
	// signature encoded as hex.
	Bip340Sig schnorr.Signature

	// TxID is a Bitcoin transaction ID. chainhash.Hash matches CLN's txid JSON
	// byte order because both parse and render txids in Bitcoin display order.
	TxID = chainhash.Hash

	// Outpoint is a Bitcoin transaction output reference of the form txid:index.
	Outpoint wire.OutPoint

	// Feerate is a CLN feerate value, such as "normal", "urgent", "253perkw",
	// or "1000perkb".
	Feerate string
)

// Hash is a raw 32-byte SHA256-style value rendered as hex without txid-style
// byte reversal. CLN uses this for payment hashes, offer IDs, merkle roots, and
// other generic hashes.
type Hash [32]byte

// Secret is a raw 32-byte secret/preimage value rendered as hex. It is not
// necessarily a secp256k1 private key.
type Secret [32]byte

// Sat is a satoshi-denominated amount.
type Sat uint64

// MSat is a millisatoshi-denominated amount.
type MSat uint64

// Amount is the default satoshi-denominated amount used for on-chain output
// descriptors.
type Amount = Sat

// AmountOrAll is either a satoshi-denominated amount or CLN's "all" literal,
// meaning all available funds for that parameter.
type AmountOrAll struct {
	Amount Amount
	All    bool
}

// MSatOrAll is either a millisatoshi-denominated amount or CLN's "all" literal.
type MSatOrAll struct {
	Amount MSat
	All    bool
}

// AmountOrAny is either a millisatoshi-denominated amount or CLN's "any"
// literal, used by invoice-like RPCs where the payer may choose the amount.
type AmountOrAny struct {
	Amount MSat
	Any    bool
}

// OutputDesc is a Bitcoin-style output descriptor: destination string to
// satoshi amount or "all". CLN uses it in withdrawal-style RPCs.
type OutputDesc map[string]AmountOrAll

// ShortChannelID is the packed BOLT 7 short_channel_id:
// 24 bits block height, 24 bits transaction index, and 16 bits output index.
type ShortChannelID uint64

// ShortChannelIDDir is a short_channel_id plus a direction bit, rendered by CLN
// as blockheightxtxindexxoutnum/0 or /1.
type ShortChannelIDDir struct {
	ShortChannelID ShortChannelID
	Direction      byte
}

// NewAmountSats returns a satoshi-denominated Amount.
func NewAmountSats(sats uint64) Amount {
	return Amount(sats)
}

// NewMSat returns a millisatoshi-denominated amount.
func NewMSat(msat uint64) MSat {
	return MSat(msat)
}

// AllAmount returns an AmountOrAll containing CLN's "all" literal.
func AllAmount() AmountOrAll {
	return AmountOrAll{All: true}
}

// AmountValue wraps a concrete satoshi-denominated amount.
func AmountValue(amount Amount) AmountOrAll {
	return AmountOrAll{Amount: amount}
}

// AnyAmount returns an AmountOrAny containing CLN's "any" literal.
func AnyAmount() AmountOrAny {
	return AmountOrAny{Any: true}
}

// AmountAnyValue wraps a concrete millisatoshi-denominated amount.
func AmountAnyValue(amount MSat) AmountOrAny {
	return AmountOrAny{Amount: amount}
}

func (a Amount) Uint64() uint64 {
	return uint64(a)
}

func (m MSat) Uint64() uint64 {
	return uint64(m)
}

func (p PubKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(p.SerializeCompressed()))
}

func (p *PubKey) UnmarshalJSON(b []byte) error {
	decoded, err := unmarshalHexString(b)
	if err != nil {
		return err
	}
	key, err := btcec.ParsePubKey(decoded)
	if err != nil {
		return fmt.Errorf("invalid pubkey: %w", err)
	}
	*p = PubKey(*key)
	return nil
}

func (p PubKey) SerializeCompressed() []byte {
	return ((*btcec.PublicKey)(&p)).SerializeCompressed()
}

func (s Signature) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(((*ecdsa.Signature)(&s)).Serialize()))
}

func (s *Signature) UnmarshalJSON(b []byte) error {
	decoded, err := unmarshalHexString(b)
	if err != nil {
		return err
	}
	sig, err := ecdsa.ParseDERSignature(decoded)
	if err != nil {
		return fmt.Errorf("invalid signature: %w", err)
	}
	*s = Signature(*sig)
	return nil
}

func (s Bip340Sig) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(((*schnorr.Signature)(&s)).Serialize()))
}

func (s *Bip340Sig) UnmarshalJSON(b []byte) error {
	decoded, err := unmarshalHexString(b)
	if err != nil {
		return err
	}
	sig, err := schnorr.ParseSignature(decoded)
	if err != nil {
		return fmt.Errorf("invalid bip340sig: %w", err)
	}
	*s = Bip340Sig(*sig)
	return nil
}

func (o Outpoint) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

func (o *Outpoint) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	outpoint, err := wire.NewOutPointFromString(raw)
	if err != nil {
		return fmt.Errorf("invalid outpoint: %w", err)
	}
	*o = Outpoint(*outpoint)
	return nil
}

func (o Outpoint) String() string {
	return wire.OutPoint(o).String()
}

func (o *Outpoint) Wire() *wire.OutPoint {
	return (*wire.OutPoint)(o)
}

func (a Amount) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint64(a))
}

func (a *Amount) UnmarshalJSON(b []byte) error {
	v, err := parseSatAmountJSON(b)
	if err != nil {
		return err
	}
	*a = Amount(v)
	return nil
}

func (m MSat) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint64(m))
}

func (m *MSat) UnmarshalJSON(b []byte) error {
	v, err := parseMSatAmountJSON(b)
	if err != nil {
		return err
	}
	*m = MSat(v)
	return nil
}

func (a AmountOrAll) MarshalJSON() ([]byte, error) {
	if a.All {
		return json.Marshal("all")
	}
	return json.Marshal(a.Amount)
}

func (a *AmountOrAll) UnmarshalJSON(b []byte) error {
	if isJSONStringLiteral(b, "all") {
		*a = AmountOrAll{All: true}
		return nil
	}

	var amount Amount
	if err := amount.UnmarshalJSON(b); err != nil {
		return err
	}
	*a = AmountOrAll{Amount: amount}
	return nil
}

func (a MSatOrAll) MarshalJSON() ([]byte, error) {
	if a.All {
		return json.Marshal("all")
	}
	return json.Marshal(a.Amount)
}

func (a *MSatOrAll) UnmarshalJSON(b []byte) error {
	if isJSONStringLiteral(b, "all") {
		*a = MSatOrAll{All: true}
		return nil
	}

	var amount MSat
	if err := amount.UnmarshalJSON(b); err != nil {
		return err
	}
	*a = MSatOrAll{Amount: amount}
	return nil
}

func (a AmountOrAny) MarshalJSON() ([]byte, error) {
	if a.Any {
		return json.Marshal("any")
	}
	return json.Marshal(a.Amount)
}

func (a *AmountOrAny) UnmarshalJSON(b []byte) error {
	if isJSONStringLiteral(b, "any") {
		*a = AmountOrAny{Any: true}
		return nil
	}

	var amount MSat
	if err := amount.UnmarshalJSON(b); err != nil {
		return err
	}
	*a = AmountOrAny{Amount: amount}
	return nil
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(h[:]))
}

func (h *Hash) UnmarshalJSON(b []byte) error {
	return unmarshal32ByteHex(b, h[:])
}

func (s Secret) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(s[:]))
}

func (s *Secret) UnmarshalJSON(b []byte) error {
	return unmarshal32ByteHex(b, s[:])
}

func (s ShortChannelID) BlockHeight() uint32 {
	return uint32(uint64(s) >> 40)
}

func (s ShortChannelID) TxIndex() uint32 {
	return uint32((uint64(s) >> 16) & 0x00ffffff)
}

func (s ShortChannelID) OutputIndex() uint16 {
	return uint16(uint64(s) & 0xffff)
}

func (s ShortChannelID) String() string {
	return fmt.Sprintf("%dx%dx%d", s.BlockHeight(), s.TxIndex(), s.OutputIndex())
}

func (s ShortChannelID) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *ShortChannelID) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	scid, err := ParseShortChannelID(raw)
	if err != nil {
		return err
	}
	*s = scid
	return nil
}

func ParseShortChannelID(raw string) (ShortChannelID, error) {
	parts := strings.Split(raw, "x")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid short_channel_id %q", raw)
	}

	blockHeight, err := parseBoundedUint(parts[0], 24, "block height")
	if err != nil {
		return 0, err
	}
	txIndex, err := parseBoundedUint(parts[1], 24, "tx index")
	if err != nil {
		return 0, err
	}
	outputIndex, err := parseBoundedUint(parts[2], 16, "output index")
	if err != nil {
		return 0, err
	}

	return ShortChannelID(blockHeight<<40 | txIndex<<16 | outputIndex), nil
}

func (s ShortChannelIDDir) String() string {
	return fmt.Sprintf("%s/%d", s.ShortChannelID.String(), s.Direction)
}

func (s ShortChannelIDDir) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *ShortChannelIDDir) UnmarshalJSON(b []byte) error {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	scidDir, err := ParseShortChannelIDDir(raw)
	if err != nil {
		return err
	}
	*s = scidDir
	return nil
}

func ParseShortChannelIDDir(raw string) (ShortChannelIDDir, error) {
	scidRaw, dirRaw, ok := strings.Cut(raw, "/")
	if !ok {
		return ShortChannelIDDir{}, fmt.Errorf("invalid short_channel_id_dir %q", raw)
	}

	scid, err := ParseShortChannelID(scidRaw)
	if err != nil {
		return ShortChannelIDDir{}, err
	}
	dir, err := parseBoundedUint(dirRaw, 1, "direction")
	if err != nil {
		return ShortChannelIDDir{}, err
	}

	return ShortChannelIDDir{ShortChannelID: scid, Direction: byte(dir)}, nil
}

func unmarshalHexString(b []byte) ([]byte, error) {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	decoded, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %w", err)
	}
	return decoded, nil
}

func unmarshal32ByteHex(b []byte, dst []byte) error {
	decoded, err := unmarshalHexString(b)
	if err != nil {
		return err
	}
	if len(decoded) != 32 {
		return fmt.Errorf("invalid length %d, want 32", len(decoded))
	}
	copy(dst, decoded)
	return nil
}

func parseBoundedUint(raw string, bits int, name string) (uint64, error) {
	v, err := strconv.ParseUint(raw, 10, bits)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", name, raw, err)
	}
	if v >= 1<<bits {
		return 0, fmt.Errorf("%s %d exceeds %d bits", name, v, bits)
	}
	return v, nil
}

func parseSatAmountJSON(b []byte) (amount uint64, err error) {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		// Maybe a JSON number
		var num json.Number
		if err := json.Unmarshal(b, &num); err != nil {
			return 0, err
		}
		raw = num.String()
	}

	if strings.HasSuffix(raw, "msat") {
		amount, err = parseMSatString(raw)
	} else if strings.HasSuffix(raw, "sat") {
		amount, err = parseSatString(raw)
	} else if strings.HasSuffix(raw, "btc") {
		amount, err = parseBtcString(raw)
	} else {
		return strconv.ParseUint(raw, 10, 64)
	}

	if amount%1000 != 0 {
		return 0, fmt.Errorf("invalid sat amount %d: has msats", amount)
	}

	amount /= 1000
	return
}

func parseMSatAmountJSON(b []byte) (uint64, error) {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		// Maybe a JSON number
		var num json.Number
		if err := json.Unmarshal(b, &num); err != nil {
			return 0, err
		}
		raw = num.String()
	}

	if strings.HasSuffix(raw, "msat") {
		return parseMSatString(raw)
	}
	if strings.HasSuffix(raw, "sat") {
		return parseSatString(raw)
	}
	if strings.HasSuffix(raw, "btc") {
		return parseBtcString(raw)
	}

	return strconv.ParseUint(raw, 10, 64)
}

// parseMSatString consumes "100msat" and returns msat integer
func parseMSatString(raw string) (uint64, error) {
	return strconv.ParseUint(strings.TrimSuffix(raw, "msat"), 10, 64)
}

// parseSatString consumes "10sat" or "10.000sat" and returns msat integer
func parseSatString(raw string) (uint64, error) {
	raw = strings.TrimSuffix(raw, "sat")

	if strings.ContainsRune(raw, '.') {
		raws := strings.Split(raw, ".")
		if len(raws) != 2 {
			return 0, fmt.Errorf("unexpected decimal string format: more then one dot separator")
		}

		if len(raws[1]) != 3 {
			return 0, fmt.Errorf("unexpected decimal string format: sat fractional part should have 3 digits")
		}

		raw = raws[0] + raws[1]
	} else {
		raw += strings.Repeat("0", 3)
	}

	return strconv.ParseUint(raw, 10, 64)
}

// parseBtcString consumes "10btc" or "0.00000010btc"  or "0.00000010000btc" and returns msat integer
func parseBtcString(raw string) (uint64, error) {
	raw = strings.TrimSuffix(raw, "btc")

	if strings.ContainsRune(raw, '.') {
		raws := strings.Split(raw, ".")
		if len(raws) != 2 {
			return 0, fmt.Errorf("unexpected decimal string format: more then one dot separator")
		}

		if len(raws[1]) != 8 && len(raws[1]) != 11 {
			return 0, fmt.Errorf("unexpected decimal string format: btc fractional part should have 8 or 11 digits")
		}

		if len(raws[1]) == 8 {
			raws[1] += strings.Repeat("0", 3)
		}

		raw = raws[0] + raws[1]
	} else {
		raw += strings.Repeat("0", 11)
	}

	return strconv.ParseUint(raw, 10, 64)
}

func isJSONStringLiteral(b []byte, literal string) bool {
	var raw string
	if err := json.Unmarshal(b, &raw); err != nil {
		return false
	}
	return raw == literal
}
