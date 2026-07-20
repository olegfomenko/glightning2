package main

import (
	"github.com/olegfomenko/glightning2/clnrpc"
	"github.com/olegfomenko/glightning2/clntypes"
)

type SettleRequest struct {
	Preimage clntypes.Secret `json:"preimage"`
}

type SettleResponse struct {
	PaymentHash clntypes.Hash `json:"payment_hash"`
}

type HTLCAcceptedResponse struct {
	Result         clnrpc.HTLCAcceptedResult `json:"result"`
	FailureMessage clntypes.Hex              `json:"failure_message,omitempty"`
	PaymentKey     *clntypes.Secret          `json:"payment_key,omitempty"`
}
