package plugin

import (
	"encoding/json"

	"github.com/olegfomenko/glightning2/clnrpc"
)

// RPCHandler handles a raw JSON-RPC request sent to a plugin method.
type RPCHandler func(json.RawMessage) (any, error)

// EventHandler handles a raw event notification sent to a plugin.
type EventHandler func(json.RawMessage) error

// HookHandler handles a Core Lightning hook request.
type HookHandler func(json.RawMessage) (any, error)

type RPCMethod struct {
	manifest ManifestRPCMethod
	handler  RPCHandler
}

type HookMethod struct {
	manifest ManifestHook
	handler  HookHandler
}

type pluginDeclarations struct {
	options       []Option
	notifications []Notification
	subscriptions []string
	rpcMethods    map[string]ManifestRPCMethod
	hooks         map[string]ManifestHook
}

// Plugin defines a Core Lightning plugin.
type Plugin struct {
	declarations *pluginDeclarations

	rpcMethods map[string]RPCHandler
	events     map[string]EventHandler
	hooks      map[string]HookHandler
}

// NewPlugin creates an empty Core Lightning plugin definition.
func NewPlugin() *Plugin {
	return &Plugin{
		declarations: &pluginDeclarations{
			rpcMethods: make(map[string]ManifestRPCMethod),
			hooks:      make(map[string]ManifestHook),
		},
		rpcMethods: make(map[string]RPCHandler),
		events:     make(map[string]EventHandler),
		hooks:      make(map[string]HookHandler),
	}
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
	p.declarations.options = append(p.declarations.options, option)
	return p
}

// AddRPCMethod adds a plugin JSON-RPC method.
func (p *Plugin) AddRPCMethod(name, description string, handler RPCHandler) *Plugin {
	return p.AddRPCMethodWithManifest(ManifestRPCMethod{
		Name:        name,
		Description: description,
	}, handler)
}

// AddRPCMethodWithManifest adds a plugin JSON-RPC method with an explicit manifest entry.
func (p *Plugin) AddRPCMethodWithManifest(method ManifestRPCMethod, handler RPCHandler) *Plugin {
	p.rpcMethods[method.Name] = handler
	p.declarations.rpcMethods[method.Name] = method
	return p
}

// SubscribeEvent subscribes to an event notification.
func (p *Plugin) SubscribeEvent(name string, handler EventHandler) *Plugin {
	p.events[name] = handler
	p.declarations.subscriptions = append(p.declarations.subscriptions, name)
	return p
}

// AddNotification adds a custom notification topic the plugin can emit.
func (p *Plugin) AddNotification(method, description string) *Plugin {
	p.declarations.notifications = append(p.declarations.notifications, Notification{
		Method:      method,
		Description: description,
	})
	return p
}

// SubscribePeerConnected subscribes to the peer_connected hook.
func (p *Plugin) SubscribePeerConnected(handler func(clnrpc.PeerConnected) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "peer_connected",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

// SubscribeRecover subscribes to the recover hook.
func (p *Plugin) SubscribeRecover(handler func(clnrpc.RecoverHook) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "recover",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

// SubscribeCommitmentRevocation subscribes to the commitment_revocation hook.
func (p *Plugin) SubscribeCommitmentRevocation(handler func(clnrpc.CommitmentRevocation) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "commitment_revocation",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

// SubscribeDBWrite subscribes to the db_write hook.
func (p *Plugin) SubscribeDBWrite(handler func(clnrpc.DBWrite) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "db_write",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

// SubscribeInvoicePayment subscribes to the invoice_payment hook.
func (p *Plugin) SubscribeInvoicePayment(handler func(clnrpc.InvoicePaymentHook) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "invoice_payment",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

// SubscribeOpenchannel subscribes to the openchannel hook.
func (p *Plugin) SubscribeOpenchannel(handler func(clnrpc.Openchannel) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "openchannel",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

// SubscribeOpenchannel2 subscribes to the openchannel2 hook.
func (p *Plugin) SubscribeOpenchannel2(handler func(clnrpc.Openchannel2) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "openchannel2",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

// SubscribeOpenchannel2Changed subscribes to the openchannel2_changed hook.
func (p *Plugin) SubscribeOpenchannel2Changed(handler func(clnrpc.Openchannel2Changed) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "openchannel2_changed",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

// SubscribeOpenchannel2Sign subscribes to the openchannel2_sign hook.
func (p *Plugin) SubscribeOpenchannel2Sign(handler func(clnrpc.Openchannel2Sign) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "openchannel2_sign",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

// SubscribeRbfChannel subscribes to the rbf_channel hook.
func (p *Plugin) SubscribeRbfChannel(handler func(clnrpc.RbfChannel) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "rbf_channel",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

// SubscribeHTLCAccepted subscribes to the htlc_accepted hook.
func (p *Plugin) SubscribeHTLCAccepted(handler func(clnrpc.HTLCAccepted) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "htlc_accepted",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

// SubscribeRPCCommand subscribes to the rpc_command hook.
func (p *Plugin) SubscribeRPCCommand(handler func(clnrpc.RPCCommand) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "rpc_command",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

// SubscribeCustommsg subscribes to the custommsg hook.
func (p *Plugin) SubscribeCustommsg(handler func(clnrpc.CustommsgHook) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "custommsg",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

// SubscribeOnionMessageRecv subscribes to the onion_message_recv hook.
func (p *Plugin) SubscribeOnionMessageRecv(handler func(clnrpc.OnionMessageRecv) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "onion_message_recv",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

// SubscribeOnionMessageRecvSecret subscribes to the onion_message_recv_secret hook.
func (p *Plugin) SubscribeOnionMessageRecvSecret(handler func(clnrpc.OnionMessageRecvSecret) (any, error)) *Plugin {
	p.SubscribeHook(
		ManifestHook{
			Name: "onion_message_recv_secret",
		},
		unmarshallAndHandle(handler),
	)
	return p
}

func (p *Plugin) SubscribeHook(manifest ManifestHook, handler HookHandler) *Plugin {
	p.hooks[manifest.Name] = handler
	p.declarations.hooks[manifest.Name] = manifest
	return p
}

// Manifest builds the plugin getmanifest response from registered handlers and metadata.
func (p *pluginDeclarations) Manifest() Manifest {
	manifest := Manifest{
		Options:       append([]Option(nil), p.options...),
		RPCMethods:    make([]ManifestRPCMethod, 0, len(p.rpcMethods)),
		Subscriptions: p.subscriptions,
		Hooks:         make([]ManifestHook, 0, len(p.hooks)),
		Notifications: append([]Notification(nil), p.notifications...),
	}

	for _, method := range p.rpcMethods {
		manifest.RPCMethods = append(manifest.RPCMethods, method)
	}

	for _, hook := range p.hooks {
		manifest.Hooks = append(manifest.Hooks, hook)
	}

	return manifest
}

func unmarshallAndHandle[T any](handler func(T) (any, error)) func(json.RawMessage) (any, error) {
	return func(message json.RawMessage) (any, error) {
		var hook T
		if err := json.Unmarshal(message, &hook); err != nil {
			return nil, err
		}
		return handler(hook)
	}
}
