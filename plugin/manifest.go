package plugin

import "encoding/json"

// OptionType is a Core Lightning plugin option type.
type OptionType string

const (
	// OptionString stores a string option value.
	OptionString OptionType = "string"

	// OptionStringConceal stores a string option value that should be concealed in logs.
	OptionStringConceal OptionType = "string-conceal"

	// OptionBool stores a boolean option value.
	OptionBool OptionType = "bool"

	// OptionInt stores an integer option value.
	OptionInt OptionType = "int"

	// OptionFlag stores a flag option that is true when present.
	OptionFlag OptionType = "flag"
)

// Option describes a command line option accepted by a plugin.
type Option struct {
	Name        string     `json:"name"`
	Type        OptionType `json:"type"`
	Default     any        `json:"default,omitempty"`
	Description string     `json:"description"`
	Category    string     `json:"category,omitempty"`
	Dynamic     bool       `json:"dynamic,omitempty"`
	Deprecated  any        `json:"deprecated,omitempty"`
	Multi       bool       `json:"multi,omitempty"`
}

// NewStringOption creates a string plugin option.
func NewStringOption(name, description, defaultValue string) Option {
	return Option{
		Name:        name,
		Type:        OptionString,
		Default:     defaultValue,
		Description: description,
	}
}

// NewStringConcealOption creates a concealed string plugin option.
func NewStringConcealOption(name, description, defaultValue string) Option {
	return Option{
		Name:        name,
		Type:        OptionStringConceal,
		Default:     defaultValue,
		Description: description,
	}
}

// NewBoolOption creates a boolean plugin option.
func NewBoolOption(name, description string, defaultValue bool) Option {
	return Option{
		Name:        name,
		Type:        OptionBool,
		Default:     defaultValue,
		Description: description,
	}
}

// NewIntOption creates an integer plugin option.
func NewIntOption(name, description string, defaultValue int64) Option {
	return Option{
		Name:        name,
		Type:        OptionInt,
		Default:     defaultValue,
		Description: description,
	}
}

// NewFlagOption creates a flag plugin option.
func NewFlagOption(name, description string) Option {
	return Option{
		Name:        name,
		Type:        OptionFlag,
		Description: description,
	}
}

// ManifestRPCMethod describes a JSON-RPC method implemented by a plugin.
type ManifestRPCMethod struct {
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	Usage           string `json:"usage"`
	LongDescription string `json:"long_description,omitempty"`
	Category        string `json:"category,omitempty"`
	Deprecated      any    `json:"deprecated,omitempty"`
}

// ManifestHook describes a Core Lightning plugin hook subscription.
//
// A hook with only Name set is marshaled as a plain string, matching the older
// glightning manifest format. Hooks with ordering or filters are marshaled as
// objects, matching the current Core Lightning manifest format.
type ManifestHook struct {
	Name    string   `json:"name"`
	Before  []string `json:"before,omitempty"`
	After   []string `json:"after,omitempty"`
	Filters []any    `json:"filters,omitempty"`
}

// MarshalJSON emits simple hooks as strings and extended hooks as objects.
func (h ManifestHook) MarshalJSON() ([]byte, error) {
	if len(h.Before) == 0 && len(h.After) == 0 && len(h.Filters) == 0 {
		return json.Marshal(h.Name)
	}

	type hook ManifestHook
	return json.Marshal(hook(h))
}

// UnmarshalJSON accepts both the simple hook name and extended hook object forms.
func (h *ManifestHook) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		*h = ManifestHook{Name: name}
		return nil
	}

	type hook ManifestHook
	var value hook
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*h = ManifestHook(value)
	return nil
}

// FeatureBits contains feature bitsets announced by a plugin.
type FeatureBits struct {
	Node    string `json:"node,omitempty"`
	Init    string `json:"init,omitempty"`
	Invoice string `json:"invoice,omitempty"`
	Channel string `json:"channel,omitempty"`
}

// Notification describes a custom JSON-RPC notification topic emitted by a plugin.
type Notification struct {
	Method      string `json:"method"`
	Description string `json:"description,omitempty"`
}

// Manifest is the response returned by a plugin's getmanifest method.
type Manifest struct {
	Options        []Option            `json:"options"`
	RPCMethods     []ManifestRPCMethod `json:"rpcmethods"`
	Dynamic        bool                `json:"dynamic"`
	Subscriptions  []string            `json:"subscriptions,omitempty"`
	Hooks          []ManifestHook      `json:"hooks,omitempty"`
	FeatureBits    FeatureBits         `json:"featurebits"`
	Notifications  []Notification      `json:"notifications,omitempty"`
	CustomMessages []uint16            `json:"custommessages,omitempty"`
	NonNumericIDs  bool                `json:"nonnumericids,omitempty"`
	CanCheck       bool                `json:"cancheck,omitempty"`
	Disable        string              `json:"disable,omitempty"`
}

// MarshalJSON keeps options and rpcmethods as arrays even when the manifest was
// constructed without NewManifest.
func (m Manifest) MarshalJSON() ([]byte, error) {
	type manifest Manifest

	out := manifest(m)
	if out.Options == nil {
		out.Options = []Option{}
	}
	if out.RPCMethods == nil {
		out.RPCMethods = []ManifestRPCMethod{}
	}

	return json.Marshal(out)
}
