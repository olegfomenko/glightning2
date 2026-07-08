package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type service struct {
	Methods       []method `json:"methods"`
	Notifications []event  `json:"notifications"`
	Hooks         []event  `json:"hooks"`
}

type method struct {
	NameRaw  string `json:"name_raw"`
	Request  field  `json:"request"`
	Response field  `json:"response"`
}

type event struct {
	Name     string `json:"name"`
	Typename string `json:"typename"`
	Request  field  `json:"request"`
	Response field  `json:"response"`
}

type field struct {
	Kind        string          `json:"kind"`
	Path        string          `json:"path"`
	Name        string          `json:"name"`
	Normalized  string          `json:"normalized"`
	Description json.RawMessage `json:"description"`
	Required    bool            `json:"required"`
	Optional    bool            `json:"optional"`
	Omitted     bool            `json:"omitted"`
	Override    *string         `json:"override"`
	Typename    string          `json:"typename"`
	Fields      []field         `json:"fields"`
	Dims        int             `json:"dims"`
	Itemtype    *field          `json:"itemtype"`
	Variants    json.RawMessage `json:"variants"`
}

type generator struct {
	defs    map[string]string
	order   []string
	imports map[string]bool
}

func (f field) enumVariants() []string {
	var variants []string
	_ = json.Unmarshal(f.Variants, &variants)
	return variants
}

func (f field) unionVariants() []field {
	var variants []field
	_ = json.Unmarshal(f.Variants, &variants)
	return variants
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: clnrpcgen <schema-json> <out-dir>")
		os.Exit(2)
	}

	schemaPath := os.Args[1]
	outDir := os.Args[2]

	data, err := os.ReadFile(schemaPath)
	if err != nil {
		fatal(err)
	}

	var svc service
	if err := json.Unmarshal(data, &svc); err != nil {
		fatal(err)
	}

	gen := &generator{
		defs:    make(map[string]string),
		imports: make(map[string]bool),
	}
	gen.collectService(svc)

	src, err := gen.render()
	if err != nil {
		fatal(err)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "clnrpc_generated.go"), src, 0o644); err != nil {
		fatal(err)
	}
}

func (g *generator) collectService(svc service) {
	for _, m := range svc.Methods {
		g.collectField(m.Request)
		g.collectField(m.Response)
	}
	for _, n := range svc.Notifications {
		g.collectField(n.Request)
		g.collectField(n.Response)
	}
	for _, h := range svc.Hooks {
		g.collectField(h.Request)
		g.collectField(h.Response)
	}
}

func (g *generator) collectField(f field) {
	if f.Omitted {
		return
	}

	switch f.Kind {
	case "composite":
		for _, child := range f.Fields {
			g.collectField(child)
		}
		g.addStruct(f)
	case "enum":
		g.addEnum(f)
	case "array":
		if f.Itemtype != nil {
			g.collectField(*f.Itemtype)
		}
	case "union":
		for _, variant := range f.unionVariants() {
			g.collectField(variant)
		}
	}
}

func (g *generator) addStruct(f field) {
	name := goNamedType(f)
	if name == "" {
		return
	}
	if _, ok := g.defs[name]; ok {
		return
	}

	var b strings.Builder
	writeComment(&b, name, f.Description)
	fmt.Fprintf(&b, "type %s struct {\n", name)
	for _, child := range f.Fields {
		if child.Omitted {
			continue
		}
		fieldName := goFieldName(child.Normalized)
		if fieldName == "" {
			continue
		}
		fieldType := g.goType(child)
		tag := jsonTag(child)
		writeComment(&b, fieldName, child.Description)
		fmt.Fprintf(&b, "\t%s %s `%s`\n", fieldName, fieldType, tag)
	}
	b.WriteString("}\n\n")

	g.defs[name] = b.String()
	g.order = append(g.order, name)
}

func (g *generator) addEnum(f field) {
	name := goNamedType(f)
	if name == "" {
		return
	}
	if _, ok := g.defs[name]; ok {
		return
	}

	var b strings.Builder
	writeComment(&b, name, f.Description)
	fmt.Fprintf(&b, "type %s string\n\n", name)
	variants := f.enumVariants()
	if len(variants) > 0 {
		b.WriteString("const (\n")
		used := map[string]bool{}
		for _, variant := range variants {
			constName := uniqueName(name+goExportedName(variant), used)
			fmt.Fprintf(&b, "\t%s %s = %q\n", constName, name, variant)
		}
		b.WriteString(")\n\n")
	}

	g.defs[name] = b.String()
	g.order = append(g.order, name)
}

func (g *generator) goType(f field) string {
	if f.Omitted {
		return "any"
	}

	switch f.Kind {
	case "primitive":
		return g.primitiveType(f.Typename)
	case "enum", "composite":
		return goNamedType(f)
	case "array":
		item := "any"
		if f.Itemtype != nil {
			item = g.goType(*f.Itemtype)
		}
		if f.Dims < 1 {
			return "[]" + item
		}
		return strings.Repeat("[]", f.Dims) + item
	case "union":
		return "any"
	default:
		return "any"
	}
}

func (g *generator) primitiveType(name string) string {
	switch name {
	case "boolean":
		return "bool"
	case "u8":
		return "uint8"
	case "u16":
		return "uint16"
	case "u32":
		return "uint32"
	case "u64":
		return "uint64"
	case "integer":
		return "int64"
	case "number", "float":
		return "float64"
	case "f32":
		return "float32"
	case "string":
		return "string"
	case "hex":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.Hex"
	case "hash":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.Hash"
	case "secret":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.Secret"
	case "pubkey":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.PubKey"
	case "txid":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.TxID"
	case "outpoint":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.Outpoint"
	case "short_channel_id":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.ShortChannelID"
	case "short_channel_id_dir":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.ShortChannelIDDir"
	case "signature":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.Signature"
	case "bip340sig":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.Bip340Sig"
	case "currency":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.Currency"
	case "feerate":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.Feerate"
	case "outputdesc":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.OutputDesc"
	case "msat":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.MSat"
	case "sat":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.Sat"
	case "msat_or_all":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.MSatOrAll"
	case "msat_or_any":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.AmountOrAny"
	case "sat_or_all":
		g.imports["glightning2/clntypes"] = true
		return "clntypes.AmountOrAll"
	case "string_map":
		return "map[string]string"
	case "json_object_or_array", "json_scalar":
		return "any"
	default:
		return "any"
	}
}

func (g *generator) render() ([]byte, error) {
	var b strings.Builder
	b.WriteString("// Code generated by cmd/clnrpcgen; DO NOT EDIT.\n\n")
	b.WriteString("package clnrpc\n\n")

	imports := make([]string, 0, len(g.imports))
	for path := range g.imports {
		imports = append(imports, path)
	}
	sort.Strings(imports)
	if len(imports) > 0 {
		b.WriteString("import (\n")
		for _, path := range imports {
			fmt.Fprintf(&b, "\t%q\n", path)
		}
		b.WriteString(")\n\n")
	}

	for _, name := range g.order {
		b.WriteString(g.defs[name])
	}

	src, err := format.Source([]byte(b.String()))
	if err != nil {
		return []byte(b.String()), err
	}
	return src, nil
}

func jsonTag(f field) string {
	name := strings.TrimSuffix(f.Name, "[]")
	if name == "" {
		name = strings.TrimSuffix(lastPathPart(f.Path), "[]")
	}
	if f.Optional {
		return fmt.Sprintf(`json:"%s,omitempty"`, name)
	}
	return fmt.Sprintf(`json:"%s"`, name)
}

func writeComment(b *strings.Builder, name string, raw json.RawMessage) {
	text := description(raw)
	if text == "" {
		return
	}
	fmt.Fprintf(b, "\t// %s %s\n", name, firstSentence(text))
}

func description(raw json.RawMessage) string {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return ""
	}
	var parts []string
	if err := json.Unmarshal(raw, &parts); err == nil {
		return strings.Join(parts, " ")
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return text
	}
	return ""
}

func firstSentence(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return ""
	}
	return text
}

func goNamedType(f field) string {
	name := strings.TrimSuffix(f.Path, "[]")
	if strings.HasSuffix(f.Typename, "Request") {
		return goExportedName(name) + "Request"
	}
	if strings.HasSuffix(f.Typename, "Response") {
		return goExportedName(name) + "Response"
	}
	return goExportedName(name)
}

func goFieldName(name string) string {
	name = strings.TrimSuffix(name, "[]")
	if name == "" {
		return ""
	}
	return goExportedName(name)
}

func goExportedName(name string) string {
	name = normalizeKnownName(name)
	parts := splitName(name)
	if len(parts) == 0 {
		return ""
	}
	var b strings.Builder
	for _, part := range parts {
		lower := strings.ToLower(part)
		if initialisms[lower] {
			b.WriteString(strings.ToUpper(lower))
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			b.WriteString(strings.ToLower(part[1:]))
		}
	}
	out := b.String()
	if out == "" {
		return ""
	}
	if out[0] >= '0' && out[0] <= '9' {
		return "N" + out
	}
	return out
}

var nonName = regexp.MustCompile(`[^A-Za-z0-9]+`)

func splitName(name string) []string {
	name = strings.ReplaceAll(name, "[]", "")
	raw := nonName.Split(name, -1)
	parts := make([]string, 0, len(raw))
	for _, part := range raw {
		for _, word := range splitCamel(part) {
			if word != "" {
				parts = append(parts, word)
			}
		}
	}
	return parts
}

func normalizeKnownName(name string) string {
	replacements := []struct {
		old string
		new string
	}{
		{"Getinfo", "GetInfo"},
		{"getinfo", "get_info"},
	}
	for _, replacement := range replacements {
		name = strings.ReplaceAll(name, replacement.old, replacement.new)
	}
	return name
}

func splitCamel(name string) []string {
	if name == "" {
		return nil
	}

	runes := []rune(name)
	var parts []string
	start := 0
	for i := 1; i < len(runes); i++ {
		prev := runes[i-1]
		cur := runes[i]
		nextLower := i+1 < len(runes) && isLower(runes[i+1])
		if (isLower(prev) || isDigit(prev)) && isUpper(cur) {
			parts = append(parts, string(runes[start:i]))
			start = i
		} else if isUpper(prev) && isUpper(cur) && nextLower {
			parts = append(parts, string(runes[start:i]))
			start = i
		}
	}
	parts = append(parts, string(runes[start:]))
	return parts
}

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
func isLower(r rune) bool { return r >= 'a' && r <= 'z' }
func isDigit(r rune) bool { return r >= '0' && r <= '9' }

func lastPathPart(path string) string {
	if idx := strings.LastIndex(path, "."); idx >= 0 {
		return path[idx+1:]
	}
	return path
}

func uniqueName(name string, used map[string]bool) string {
	if name == "" {
		name = "Value"
	}
	base := name
	for i := 2; used[name]; i++ {
		name = fmt.Sprintf("%s%d", base, i)
	}
	used[name] = true
	return name
}

var initialisms = map[string]bool{
	"api":  true,
	"bip":  true,
	"cln":  true,
	"csv":  true,
	"db":   true,
	"hex":  true,
	"htlc": true,
	"id":   true,
	"json": true,
	"msat": true,
	"psbt": true,
	"rpc":  true,
	"scid": true,
	"tlv":  true,
	"tx":   true,
	"txid": true,
	"uri":  true,
	"url":  true,
	"utxo": true,
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
