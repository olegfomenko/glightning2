package plugin

import (
	"encoding/json"
	"testing"
)

func TestManifestBuilder(t *testing.T) {
	manifest := NewManifestBuilder().
		StringOption("flag1", "Flag 1 option (string)", "string").
		BoolOption("flag2", "Flag 2 option (bool)", true).
		RPCMethod("get_user", "get_user --name=Oleg", "Get user entry").
		RPCMethod("save_user", "save_user --name=Oleg --age=23 --city=Kyiv", "Save user entry").
		Subscribe("connect").
		Hook("db_write").
		Hook("htlc_accepted").
		Hook("custommsg").
		CustomMessage(11008).
		Build()

	got, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}

	want := `{"options":[{"name":"flag1","type":"string","default":"string","description":"Flag 1 option (string)"},{"name":"flag2","type":"bool","default":true,"description":"Flag 2 option (bool)"}],"rpcmethods":[{"name":"get_user","description":"Get user entry","usage":"get_user --name=Oleg"},{"name":"save_user","description":"Save user entry","usage":"save_user --name=Oleg --age=23 --city=Kyiv"}],"dynamic":true,"subscriptions":["connect"],"hooks":["db_write","htlc_accepted","custommsg"],"featurebits":{},"custommessages":[11008]}`
	if string(got) != want {
		t.Fatalf("manifest JSON mismatch\nwant: %s\n got: %s", want, got)
	}
}
