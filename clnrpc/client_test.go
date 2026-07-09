package clnrpc

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestClientGetInfoWithLightningD(t *testing.T) {
	lightningd, err := exec.LookPath("lightningd")
	if err != nil {
		t.Skip("lightningd is not installed")
	}

	bitcoind, err := exec.LookPath("bitcoind")
	if err != nil {
		t.Skip("bitcoind is not installed")
	}

	bitcoinDir, err := os.MkdirTemp("/tmp", "btc")
	if err != nil {
		t.Fatalf("create bitcoind temp dir: %v", err)
	}

	lightningDir, err := os.MkdirTemp("/tmp", "cln")
	if err != nil {
		t.Fatalf("create lightningd temp dir: %v", err)
	}

	t.Cleanup(func() {
		_ = os.RemoveAll(bitcoinDir)
		_ = os.RemoveAll(lightningDir)
	})

	var bitcoinLogs bytes.Buffer
	bitcoin := exec.Command(
		bitcoind,
		"-regtest",
		"-datadir="+bitcoinDir,
		"-server",
		"-listen=0",
		"-rpcuser=user",
		"-rpcpassword=password",
		"-fallbackfee=0.0001",
		"-printtoconsole",
	)
	bitcoin.Stdout = &bitcoinLogs
	bitcoin.Stderr = &bitcoinLogs

	if err := bitcoin.Start(); err != nil {
		t.Fatalf("start bitcoind: %v", err)
	}
	defer func() {
		_ = bitcoin.Process.Kill()
	}()

	time.Sleep(5 * time.Second)

	socketPath := filepath.Join(lightningDir, "regtest", "lightning-rpc")

	var logs bytes.Buffer
	cmd := exec.Command(
		lightningd,
		"--lightning-dir="+lightningDir,
		"--regtest",
		"--autolisten=false",
		"--developer",
		"--dev-no-version-checks",
		"--bitcoin-rpcconnect=127.0.0.1",
		"--bitcoin-rpcuser=user",
		"--bitcoin-rpcpassword=password",
		"--log-file=-",
		"--log-level=debug",
	)
	cmd.Stdout = &logs
	cmd.Stderr = &logs

	if err := cmd.Start(); err != nil {
		t.Fatalf("start lightningd: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()

	time.Sleep(5 * time.Second)

	if _, err := os.Stat(socketPath); err != nil {
		t.Skipf("lightningd did not create RPC socket: %v\nbitcoind:\n%s\nlightningd:\n%s", err, bitcoinLogs.String(), logs.String())
	}

	callCtx, cancelCall := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCall()
	info, err := NewClient(socketPath).GetInfo(callCtx)
	if err != nil {
		t.Fatalf("getinfo failed: %v\n%s", err, logs.String())
	}

	if info == nil {
		t.Fatal("getinfo returned nil response")
	}
	if len(info.ID.SerializeCompressed()) != 33 {
		t.Fatalf("getinfo returned empty node id: %#v", info)
	}
	if info.Network != "regtest" {
		t.Fatalf("network = %q, want regtest", info.Network)
	}
}
