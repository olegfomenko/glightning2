package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/olegfomenko/glightning2/clnrpc"
	"github.com/olegfomenko/glightning2/plugin"
)

const (
	maxConcurrentRequests = 128

	// WireIncorrectOrUnknownPaymentDetails constant for fail responses of htlc_accepted hook.
	// Taken from lightning/wire/onion_wiregen.h
	WireIncorrectOrUnknownPaymentDetails = "400f"
)

// HoldInvoicePlugin holds final-hop HTLCs until their preimage is supplied.
// Plugin is embedded so the SDK's builder, lifecycle, client, and logger APIs
// are promoted onto the composed type.
type HoldInvoicePlugin struct {
	*plugin.Plugin
}

// NewHoldInvoicePlugin builds the example hold-invoice plugin.
func NewHoldInvoicePlugin() *HoldInvoicePlugin {
	p := &HoldInvoicePlugin{Plugin: plugin.NewPlugin()}

	// A held hook occupies one request worker. Keep spare capacity so the
	// settlement RPC can still run while multiple HTLCs are waiting.
	p.SetMaxConcurrentRequests(maxConcurrentRequests)

	p.AddIntOption(
		"holdinvoice-poll-seconds",
		"Seconds between datastore checks while an HTLC is held",
		5,
	)
	p.AddRPCMethodWithManifest(plugin.ManifestRPCMethod{
		Name:        "holdinvoice-settle",
		Description: "Set the preimage for a held invoice; its payment hash is derived automatically",
		Usage:       "preimage",
	}, p.handleRPCSettle)

	p.SubscribeHTLCAccepted(p.handleHTLCAccepted)

	return p
}

func (p *HoldInvoicePlugin) handleRPCSettle(ctx context.Context, params json.RawMessage) (json.RawMessage, error) {
	var request SettleRequest
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, fmt.Errorf("decode settle request: %w", err)
	}

	key := HoldInvoiceEntryKey(sha256.Sum256(request.Preimage[:]))
	entry := HoldInvoiceEntry{Preimage: &request.Preimage}

	if err := p.DataStore().Save(ctx, key, &entry, clnrpc.DatastoreModeMustReplace); err != nil {
		return nil, fmt.Errorf("update held invoice: %w", err)
	}

	return json.Marshal(SettleResponse{PaymentHash: sha256.Sum256(request.Preimage[:])})
}

func (p *HoldInvoicePlugin) handleHTLCAccepted(ctx context.Context, hook clnrpc.HTLCAccepted) (json.RawMessage, error) {
	if !isFinalDestination(hook) {
		return json.Marshal(HTLCAcceptedResponse{Result: clnrpc.HTLCAcceptedResultContinue})
	}

	key := HoldInvoiceEntryKey(hook.HTLC.PaymentHash)
	p.GetLogger().Info(ctx, fmt.Sprintf("holding %s", key[1]))

	pollInterval, err := p.PollInterval()
	if err != nil {
		p.GetLogger().Error(ctx, fmt.Sprintf("holding %s failed: %v", key[1], err))
		return json.Marshal(HTLCAcceptedResponse{Result: clnrpc.HTLCAcceptedResultContinue})
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		// Get current block
		info, err := p.GetClient().GetInfo(ctx)
		if err != nil {
			p.GetLogger().Error(ctx, fmt.Sprintf("holding %s failed: %v", key[1], err))
			return json.Marshal(HTLCAcceptedResponse{Result: clnrpc.HTLCAcceptedResultContinue})
		}

		// Check if CLTV expired
		if uint64(info.Blockheight) >= uint64(hook.HTLC.CltvExpiry) {
			p.GetLogger().Info(ctx, fmt.Sprintf("holding %s failed: rejecting by CLTV", key[1]))

			return json.Marshal(HTLCAcceptedResponse{
				Result:         clnrpc.HTLCAcceptedResultFail,
				FailureMessage: WireIncorrectOrUnknownPaymentDetails,
			})
		}

		// Wait ticker
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}

		// Check for entry
		entry, err := p.DataStore().Get(ctx, key)
		if err != nil && !errors.Is(err, plugin.ErrDatastoreNotFound) {
			p.GetLogger().Error(ctx, fmt.Sprintf("holding %s failed: %v", key[1], err))
			return json.Marshal(HTLCAcceptedResponse{Result: clnrpc.HTLCAcceptedResultContinue})
		}

		p.GetLogger().Info(ctx, fmt.Sprintf("holding %s resolved", key[1]))

		return json.Marshal(HTLCAcceptedResponse{
			Result:     clnrpc.HTLCAcceptedResultResolve,
			PaymentKey: entry.Preimage,
		})
	}
}

func (p *HoldInvoicePlugin) DataStore() *plugin.DatastoreClient[HoldInvoiceEntry] {
	return plugin.GetDatastore[HoldInvoiceEntry](p.GetClient())
}

func (p *HoldInvoicePlugin) PollInterval() (time.Duration, error) {
	seconds, err := p.GetIntOption("holdinvoice-poll-seconds")
	if err != nil {
		return 0, err
	}

	return time.Duration(*seconds) * time.Second, nil
}

func isFinalDestination(hook clnrpc.HTLCAccepted) bool {
	return hook.ForwardTo == nil &&
		hook.Onion.ShortChannelID == nil &&
		hook.Onion.NextNodeID == nil
}
