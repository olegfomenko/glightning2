package plugin

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"testing"

	"github.com/olegfomenko/glightning2/clnrpc"
	"golang.org/x/exp/jsonrpc2"
)

func TestPluginConfiguration(t *testing.T) {
	getUser := func(context.Context, json.RawMessage) (json.RawMessage, error) {
		return nil, nil
	}
	onConnect := func(context.Context, json.RawMessage) error {
		return nil
	}
	onDBWrite := func(context.Context, clnrpc.DBWrite) (json.RawMessage, error) {
		return nil, nil
	}

	plugin := NewPlugin().
		AddStringOption("flag1", "Flag 1 option (string)", "string").
		AddBoolOption("flag2", "Flag 2 option (bool)", true).
		AddRPCMethod("get_user", "Get user entry", getUser).
		AddRPCMethod("save_user", "Save user entry", func(context.Context, json.RawMessage) (json.RawMessage, error) {
			return nil, nil
		}).
		SubscribeEvent("connect", onConnect).
		SubscribeDBWrite(onDBWrite).
		SubscribeHTLCAccepted(func(context.Context, clnrpc.HTLCAccepted) (json.RawMessage, error) {
			return nil, nil
		}).
		SubscribeCustommsg(func(context.Context, clnrpc.CustommsgHook) (json.RawMessage, error) {
			return nil, nil
		}).
		AddNotification("user_saved", "User was saved")

	manifest := plugin.declarations.Manifest()
	sort.Slice(manifest.RPCMethods, func(i, j int) bool {
		return manifest.RPCMethods[i].Name < manifest.RPCMethods[j].Name
	})
	sort.Slice(manifest.Hooks, func(i, j int) bool {
		return manifest.Hooks[i].Name < manifest.Hooks[j].Name
	})

	got, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}

	want := `{"options":[{"name":"flag1","type":"string","default":"string","description":"Flag 1 option (string)"},{"name":"flag2","type":"bool","default":true,"description":"Flag 2 option (bool)"}],"rpcmethods":[{"name":"get_user","description":"Get user entry","usage":""},{"name":"save_user","description":"Save user entry","usage":""}],"dynamic":false,"subscriptions":["connect"],"hooks":["custommsg","db_write","htlc_accepted"],"featurebits":{},"notifications":[{"method":"user_saved","description":"User was saved"}]}`
	if string(got) != want {
		t.Fatalf("manifest JSON mismatch\nwant: %s\n got: %s", want, got)
	}
}

func TestPluginIntegration(t *testing.T) {
	getUser := func(context.Context, json.RawMessage) (json.RawMessage, error) {
		return nil, nil
	}
	onConnect := func(context.Context, json.RawMessage) error {
		return nil
	}
	onDBWrite := func(context.Context, clnrpc.DBWrite) (json.RawMessage, error) {
		return nil, nil
	}

	plugin := NewPlugin().
		AddStringOption("flag1", "Flag 1 option (string)", "string").
		AddBoolOption("flag2", "Flag 2 option (bool)", true).
		AddRPCMethod("get_user", "Get user entry", getUser).
		AddRPCMethod("save_user", "Save user entry", func(context.Context, json.RawMessage) (json.RawMessage, error) {
			return nil, nil
		}).
		SubscribeEvent("connect", onConnect).
		SubscribeDBWrite(onDBWrite).
		SubscribeHTLCAccepted(func(context.Context, clnrpc.HTLCAccepted) (json.RawMessage, error) {
			return nil, nil
		}).
		SubscribeCustommsg(func(context.Context, clnrpc.CustommsgHook) (json.RawMessage, error) {
			return nil, nil
		}).
		AddNotification("user_saved", "User was saved")
	if plugin.GetClient() != nil {
		t.Fatal("client was initialized before init")
	}

	toPluginPipeR, toPluginPipeW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	toClnPipeR, toClnPipeW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		t.Log("Starting plugin...")
		done <- plugin.Start(ctx, toPluginPipeR, toClnPipeW)
	}()
	t.Cleanup(func() {
		cancel()
		_ = toPluginPipeR.Close()
		_ = toPluginPipeW.Close()
		_ = toClnPipeR.Close()
		_ = toClnPipeW.Close()
		<-done
		t.Log("Plugin finished.")
	})

	// Response reader
	scanner := prepareScanner(toClnPipeR, DefaultMaxIntakeBuffer)

	// Send getmanifest request
	getManifestRequest := []byte(`{"jsonrpc":"2.0","id":"manifest-id","method":"getmanifest","params":{"allow-deprecated-apis":false}}` + "\n\n")
	if _, err := toPluginPipeW.Write(getManifestRequest); err != nil {
		t.Fatal(err)
	}

	// Get response
	if !scanner.Scan() {
		t.Fatalf("reading getmanifest response: %v", scanner.Err())
	}

	// Parse response
	message, err := jsonrpc2.DecodeMessage(scanner.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	response, ok := message.(*jsonrpc2.Response)
	if !ok {
		t.Fatalf("getmanifest returned %T", message)
	}
	if response.ID.Raw() != "manifest-id" {
		t.Fatalf("getmanifest response ID: want manifest-id, got %v", response.ID.Raw())
	}
	if response.Error != nil {
		t.Fatalf("getmanifest failed: %v", response.Error)
	}
	var manifest Manifest
	if err := json.Unmarshal(response.Result, &manifest); err != nil {
		t.Fatal(err)
	}

	// Assertion
	sort.Slice(manifest.RPCMethods, func(i, j int) bool {
		return manifest.RPCMethods[i].Name < manifest.RPCMethods[j].Name
	})
	sort.Slice(manifest.Hooks, func(i, j int) bool {
		return manifest.Hooks[i].Name < manifest.Hooks[j].Name
	})
	gotManifest, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	wantManifest := `{"options":[{"name":"flag1","type":"string","default":"string","description":"Flag 1 option (string)"},{"name":"flag2","type":"bool","default":true,"description":"Flag 2 option (bool)"}],"rpcmethods":[{"name":"get_user","description":"Get user entry","usage":""},{"name":"save_user","description":"Save user entry","usage":""}],"dynamic":false,"subscriptions":["connect"],"hooks":["custommsg","db_write","htlc_accepted"],"featurebits":{},"notifications":[{"method":"user_saved","description":"User was saved"}]}`
	if string(gotManifest) != wantManifest {
		t.Fatalf("getmanifest result mismatch\nwant: %s\n got: %s", wantManifest, gotManifest)
	}

	// Send init request
	initRequest := []byte(`{"jsonrpc":"2.0","id":"init-id","method":"init","params":{"options":{"flag1":"configured","flag2":false},"configuration":{"lightning-dir":"/tmp/lightning/regtest","rpc-file":"lightning-rpc","startup":true,"network":"regtest","feature_set":{"init":"02"},"proxy":{"type":"ipv4","address":"127.0.0.1","port":9050},"torv3-enabled":true,"always_use_proxy":false}}}` + "\n\n")
	if _, err := toPluginPipeW.Write(initRequest); err != nil {
		t.Fatal(err)
	}

	// Receive response
	if !scanner.Scan() {
		t.Fatalf("reading init response: %v", scanner.Err())
	}

	// Parse response
	message, err = jsonrpc2.DecodeMessage(scanner.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	response, ok = message.(*jsonrpc2.Response)
	if !ok {
		t.Fatalf("init returned %T", message)
	}
	if response.ID.Raw() != "init-id" {
		t.Fatalf("init response ID: want init-id, got %v", response.ID.Raw())
	}
	if response.Error != nil {
		t.Fatalf("init failed: %v", response.Error)
	}
	if string(response.Result) != "{}" {
		t.Fatalf("init result: want {}, got %s", response.Result)
	}

	// Assertion

	gotConfiguration, err := json.Marshal(plugin.GetConfiguration())
	if err != nil {
		t.Fatal(err)
	}
	wantConfiguration := `{"lightning-dir":"/tmp/lightning/regtest","rpc-file":"lightning-rpc","startup":true,"network":"regtest","feature_set":{"init":"02"},"proxy":{"type":"ipv4","address":"127.0.0.1","port":9050},"torv3-enabled":true,"always_use_proxy":false}`
	if string(gotConfiguration) != wantConfiguration {
		t.Fatalf("configuration mismatch\nwant: %s\n got: %s", wantConfiguration, gotConfiguration)
	}

	client := plugin.GetClient()
	if client == nil {
		t.Fatal("client was not initialized")
	}

	stringValue, err := plugin.GetStringOption("flag1")
	if err != nil || stringValue == nil || *stringValue != "configured" {
		t.Fatalf("flag1: value=%v err=%v", stringValue, err)
	}

	boolValue, err := plugin.GetBoolOption("flag2")
	if err != nil || boolValue == nil || *boolValue {
		t.Fatalf("flag2: value=%v err=%v", boolValue, err)
	}

}
