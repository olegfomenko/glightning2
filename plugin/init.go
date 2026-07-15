package plugin

import (
	"context"
	"encoding/json"
)

// InitRequest contains the plugin options and node configuration supplied by
// Core Lightning during plugin initialization.
type InitRequest struct {
	Options       map[string]json.RawMessage `json:"options"`
	Configuration Configuration              `json:"configuration"`
}

// Configuration contains the Core Lightning environment supplied to a plugin
// by the init request.
type Configuration struct {
	LightningDir   string      `json:"lightning-dir"`
	RPCFile        string      `json:"rpc-file"`
	Startup        bool        `json:"startup"`
	Network        string      `json:"network"`
	FeatureSet     FeatureBits `json:"feature_set"`
	Proxy          *Proxy      `json:"proxy,omitempty"`
	TorV3Enabled   *bool       `json:"torv3-enabled,omitempty"`
	AlwaysUseProxy *bool       `json:"always_use_proxy,omitempty"`
}

// Proxy contains the proxy connection supplied by Core Lightning.
type Proxy struct {
	Type    string `json:"type"`
	Address string `json:"address"`
	Port    int64  `json:"port"`
}

// GetConfiguration returns a copy of the Core Lightning configuration received
// during init. It returns nil before init is processed.
func (p *Plugin) GetConfiguration() Configuration {
	cfg := p.configuration

	if p.configuration.AlwaysUseProxy != nil {
		value := *p.configuration.AlwaysUseProxy
		cfg.AlwaysUseProxy = &value
	}

	if p.configuration.TorV3Enabled != nil {
		value := *p.configuration.TorV3Enabled
		cfg.TorV3Enabled = &value
	}

	if p.configuration.Proxy != nil {
		value := Proxy{
			Type:    p.configuration.Proxy.Type,
			Address: p.configuration.Proxy.Address,
			Port:    p.configuration.Proxy.Port,
		}
		cfg.Proxy = &value
	}

	return cfg
}

func (p *Plugin) registerLifecycleMethods() {
	p.requestHandlers["getmanifest"] = p.handleGetManifest
	p.requestHandlers["init"] = p.handleInit
}

func (p *Plugin) handleGetManifest(context.Context, json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(p.declarations.Manifest())
}

func (p *Plugin) handleInit(_ context.Context, params json.RawMessage) (json.RawMessage, error) {
	var request InitRequest
	if err := json.Unmarshal(params, &request); err != nil {
		return nil, err
	}

	p.configuration = request.Configuration
	p.optionValues = request.Options
	return json.RawMessage(`{}`), nil
}

// GetStringOption returns the value of a string option.
func (p *Plugin) GetStringOption(name string) (*string, error) {
	raw, ok := p.optionValues[name]
	if !ok {
		return nil, nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

// GetStringArrayOption returns the value of a multi string option.
func (p *Plugin) GetStringArrayOption(name string) (*[]string, error) {
	raw, ok := p.optionValues[name]
	if !ok {
		return nil, nil
	}

	var value []string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

// GetStringConcealOption returns the value of a concealed string option.
func (p *Plugin) GetStringConcealOption(name string) (*string, error) {
	raw, ok := p.optionValues[name]
	if !ok {
		return nil, nil
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

// GetStringConcealArrayOption returns the value of a multi concealed string option.
func (p *Plugin) GetStringConcealArrayOption(name string) (*[]string, error) {
	raw, ok := p.optionValues[name]
	if !ok {
		return nil, nil
	}

	var value []string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

// GetBoolOption returns the value of a boolean option.
func (p *Plugin) GetBoolOption(name string) (*bool, error) {
	raw, ok := p.optionValues[name]
	if !ok {
		return nil, nil
	}

	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

// GetIntOption returns the value of an integer option.
func (p *Plugin) GetIntOption(name string) (*int64, error) {
	raw, ok := p.optionValues[name]
	if !ok {
		return nil, nil
	}

	var value int64
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

// GetIntArrayOption returns the value of a multi integer option.
func (p *Plugin) GetIntArrayOption(name string) (*[]int64, error) {
	raw, ok := p.optionValues[name]
	if !ok {
		return nil, nil
	}

	var value []int64
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

// GetFlagOption reports whether a flag option was set.
func (p *Plugin) GetFlagOption(name string) (*bool, error) {
	raw, ok := p.optionValues[name]
	if !ok {
		return nil, nil
	}

	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	return &value, nil
}
