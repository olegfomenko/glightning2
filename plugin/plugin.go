package plugin

import (
	"encoding/json"

	"github.com/olegfomenko/glightning2/clnrpc"
)

// RPCHandler handles a raw JSON-RPC request sent to a plugin method.
type RPCHandler func(json.RawMessage) (any, error)

// EventHandler handles a raw event notification sent to a plugin.
type EventHandler func(json.RawMessage) error

// HookHandler handles a typed Core Lightning hook request.
type HookHandler[T any] func(T) (any, error)

type rpcMethod struct {
	manifest RPCMethod
	handler  RPCHandler
}

type eventSubscription struct {
	name    string
	handler EventHandler
}

type hookSubscription struct {
	name    string
	handler any
}

// Plugin defines a Core Lightning plugin.
type Plugin struct {
	options       []Option
	rpcMethods    []rpcMethod
	events        []eventSubscription
	hooks         []hookSubscription
	notifications []Notification
}

// NewPlugin creates an empty Core Lightning plugin definition.
func NewPlugin() *Plugin {
	return &Plugin{}
}

// AddStringOption adds a string option to the plugin manifest.
func (p *Plugin) AddStringOption(name, description, defaultValue string) *Plugin {
	return p.AddOption(NewStringOption(name, description, defaultValue))
}

// AddStringConcealOption adds a concealed string option to the plugin manifest.
func (p *Plugin) AddStringConcealOption(name, description, defaultValue string) *Plugin {
	return p.AddOption(NewStringConcealOption(name, description, defaultValue))
}

// AddBoolOption adds a boolean option to the plugin manifest.
func (p *Plugin) AddBoolOption(name, description string, defaultValue bool) *Plugin {
	return p.AddOption(NewBoolOption(name, description, defaultValue))
}

// AddIntOption adds an integer option to the plugin manifest.
func (p *Plugin) AddIntOption(name, description string, defaultValue int64) *Plugin {
	return p.AddOption(NewIntOption(name, description, defaultValue))
}

// AddFlagOption adds a flag option to the plugin manifest.
func (p *Plugin) AddFlagOption(name, description string) *Plugin {
	return p.AddOption(NewFlagOption(name, description))
}

// AddOption adds an option to the plugin manifest.
func (p *Plugin) AddOption(option Option) *Plugin {
	p.options = append(p.options, option)
	return p
}

// AddRPCMethod adds a plugin JSON-RPC method.
func (p *Plugin) AddRPCMethod(name, description string, handler RPCHandler) *Plugin {
	return p.AddRPCMethodWithManifest(RPCMethod{
		Name:        name,
		Description: description,
	}, handler)
}

// AddRPCMethodWithManifest adds a plugin JSON-RPC method with an explicit manifest entry.
func (p *Plugin) AddRPCMethodWithManifest(method RPCMethod, handler RPCHandler) *Plugin {
	p.rpcMethods = append(p.rpcMethods, rpcMethod{
		manifest: method,
		handler:  handler,
	})
	return p
}

// SubscribeEvent subscribes to an event notification.
func (p *Plugin) SubscribeEvent(name string, handler EventHandler) *Plugin {
	p.events = append(p.events, eventSubscription{
		name:    name,
		handler: handler,
	})
	return p
}

// AddNotification adds a custom notification topic the plugin can emit.
func (p *Plugin) AddNotification(method, description string) *Plugin {
	p.notifications = append(p.notifications, Notification{
		Method:      method,
		Description: description,
	})
	return p
}

// SubscribePeerConnected subscribes to the peer_connected hook.
func (p *Plugin) SubscribePeerConnected(handler HookHandler[clnrpc.PeerConnected]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "peer_connected",
		handler: handler,
	})
	return p
}

// SubscribeRecover subscribes to the recover hook.
func (p *Plugin) SubscribeRecover(handler HookHandler[clnrpc.RecoverHook]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "recover",
		handler: handler,
	})
	return p
}

// SubscribeCommitmentRevocation subscribes to the commitment_revocation hook.
func (p *Plugin) SubscribeCommitmentRevocation(handler HookHandler[clnrpc.CommitmentRevocation]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "commitment_revocation",
		handler: handler,
	})
	return p
}

// SubscribeDBWrite subscribes to the db_write hook.
func (p *Plugin) SubscribeDBWrite(handler HookHandler[clnrpc.DBWrite]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "db_write",
		handler: handler,
	})
	return p
}

// SubscribeInvoicePayment subscribes to the invoice_payment hook.
func (p *Plugin) SubscribeInvoicePayment(handler HookHandler[clnrpc.InvoicePaymentHook]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "invoice_payment",
		handler: handler,
	})
	return p
}

// SubscribeOpenchannel subscribes to the openchannel hook.
func (p *Plugin) SubscribeOpenchannel(handler HookHandler[clnrpc.Openchannel]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "openchannel",
		handler: handler,
	})
	return p
}

// SubscribeOpenchannel2 subscribes to the openchannel2 hook.
func (p *Plugin) SubscribeOpenchannel2(handler HookHandler[clnrpc.Openchannel2]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "openchannel2",
		handler: handler,
	})
	return p
}

// SubscribeOpenchannel2Changed subscribes to the openchannel2_changed hook.
func (p *Plugin) SubscribeOpenchannel2Changed(handler HookHandler[clnrpc.Openchannel2Changed]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "openchannel2_changed",
		handler: handler,
	})
	return p
}

// SubscribeOpenchannel2Sign subscribes to the openchannel2_sign hook.
func (p *Plugin) SubscribeOpenchannel2Sign(handler HookHandler[clnrpc.Openchannel2Sign]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "openchannel2_sign",
		handler: handler,
	})
	return p
}

// SubscribeRbfChannel subscribes to the rbf_channel hook.
func (p *Plugin) SubscribeRbfChannel(handler HookHandler[clnrpc.RbfChannel]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "rbf_channel",
		handler: handler,
	})
	return p
}

// SubscribeHTLCAccepted subscribes to the htlc_accepted hook.
func (p *Plugin) SubscribeHTLCAccepted(handler HookHandler[clnrpc.HTLCAccepted]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "htlc_accepted",
		handler: handler,
	})
	return p
}

// SubscribeRPCCommand subscribes to the rpc_command hook.
func (p *Plugin) SubscribeRPCCommand(handler HookHandler[clnrpc.RPCCommand]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "rpc_command",
		handler: handler,
	})
	return p
}

// SubscribeCustommsg subscribes to the custommsg hook.
func (p *Plugin) SubscribeCustommsg(handler HookHandler[clnrpc.CustommsgHook]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "custommsg",
		handler: handler,
	})
	return p
}

// SubscribeOnionMessageRecv subscribes to the onion_message_recv hook.
func (p *Plugin) SubscribeOnionMessageRecv(handler HookHandler[clnrpc.OnionMessageRecv]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "onion_message_recv",
		handler: handler,
	})
	return p
}

// SubscribeOnionMessageRecvSecret subscribes to the onion_message_recv_secret hook.
func (p *Plugin) SubscribeOnionMessageRecvSecret(handler HookHandler[clnrpc.OnionMessageRecvSecret]) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    "onion_message_recv_secret",
		handler: handler,
	})
	return p
}

func (p *Plugin) SubscribeHook(name string, handler any) *Plugin {
	p.hooks = append(p.hooks, hookSubscription{
		name:    name,
		handler: handler,
	})
	return p
}

// Manifest builds the plugin getmanifest response from registered handlers and metadata.
func (p *Plugin) Manifest() Manifest {
	manifest := Manifest{
		Options:       append([]Option(nil), p.options...),
		RPCMethods:    make([]RPCMethod, 0, len(p.rpcMethods)),
		Subscriptions: make([]string, 0, len(p.events)),
		Hooks:         make([]Hook, 0, len(p.hooks)),
		Notifications: append([]Notification(nil), p.notifications...),
	}

	for _, method := range p.rpcMethods {
		manifest.RPCMethods = append(manifest.RPCMethods, method.manifest)
	}

	for _, event := range p.events {
		manifest.Subscriptions = append(manifest.Subscriptions, event.name)
	}

	for _, hook := range p.hooks {
		manifest.Hooks = append(manifest.Hooks, NewHook(hook.name))
	}

	return manifest
}
