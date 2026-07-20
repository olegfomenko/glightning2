package main

import (
	"encoding/hex"
	"github.com/olegfomenko/glightning2/clntypes"
)

const datastoreNamespace = "holdinvoice"

// HoldInvoiceEntry is persisted under holdinvoice/<payment-hash>.
type HoldInvoiceEntry struct {
	Preimage *clntypes.Secret `json:"preimage,omitempty"`
}

func HoldInvoiceEntryKey(hash clntypes.Hash) []string {
	return []string{datastoreNamespace, hex.EncodeToString(hash[:])}
}
