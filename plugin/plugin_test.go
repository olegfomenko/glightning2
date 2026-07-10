package plugin

import (
	"encoding/json"
	"testing"

	"github.com/olegfomenko/glightning2/clnrpc"
)

func TestPluginManifest(t *testing.T) {
	getUser := func(json.RawMessage) (any, error) {
		return nil, nil
	}
	onConnect := func(json.RawMessage) error {
		return nil
	}
	onDBWrite := func(clnrpc.DBWrite) (any, error) {
		return nil, nil
	}

	plugin := NewPlugin().
		AddStringOption("flag1", "Flag 1 option (string)", "string").
		AddBoolOption("flag2", "Flag 2 option (bool)", true).
		AddRPCMethod("get_user", "Get user entry", getUser).
		AddRPCMethod("save_user", "Save user entry", func(json.RawMessage) (any, error) {
			return nil, nil
		}).
		SubscribeEvent("connect", onConnect).
		SubscribeDBWrite(onDBWrite).
		SubscribeHTLCAccepted(func(clnrpc.HTLCAccepted) (any, error) {
			return nil, nil
		}).
		SubscribeCustommsg(func(clnrpc.CustommsgHook) (any, error) {
			return nil, nil
		}).
		AddNotification("user_saved", "User was saved")

	got, err := json.Marshal(plugin.Manifest())
	if err != nil {
		t.Fatal(err)
	}

	want := `{"options":[{"name":"flag1","type":"string","default":"string","description":"Flag 1 option (string)"},{"name":"flag2","type":"bool","default":true,"description":"Flag 2 option (bool)"}],"rpcmethods":[{"name":"get_user","description":"Get user entry","usage":""},{"name":"save_user","description":"Save user entry","usage":""}],"dynamic":false,"subscriptions":["connect"],"hooks":["db_write","htlc_accepted","custommsg"],"featurebits":{},"notifications":[{"method":"user_saved","description":"User was saved"}]}`
	if string(got) != want {
		t.Fatalf("manifest JSON mismatch\nwant: %s\n got: %s", want, got)
	}
}
