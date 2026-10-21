package cfg

import (
	"path/filepath"
	"testing"
)

// TestCheckDirsDisjoint guards the one misconfiguration that can destroy the
// data this app exists to protect: the live, unlocked, append-only databases
// ending up under a sync daemon's watch.
func TestCheckDirsDisjoint(t *testing.T) {
	cases := []struct {
		name       string
		data, sync string
		wantErr    bool
	}{
		{"separate trees", "/home/x/.pbstudy", "/home/x/iCloud/pbstudy", false},
		{"same dir", "/home/x/.pbstudy", "/home/x/.pbstudy", true},
		{"sync inside data", "/home/x/.pbstudy", "/home/x/.pbstudy/sync", true},
		{"data inside sync", "/home/x/iCloud/pb/data", "/home/x/iCloud/pb", true},
		// The check compares whole path segments, so a directory whose name
		// merely begins with the other's is not nested.
		{"shared prefix only", "/home/x/pbstudy", "/home/x/pbstudy-sync", false},
		{"trailing slash", "/home/x/.pbstudy/", "/home/x/sync/", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := checkDirsDisjoint(filepath.FromSlash(c.data), filepath.FromSlash(c.sync))
			if (err != nil) != c.wantErr {
				t.Errorf("checkDirsDisjoint(%q, %q) error = %v, wantErr %v",
					c.data, c.sync, err, c.wantErr)
			}
		})
	}
}

// TestListenAddress pins the loopback default and the opt-in to the network:
// an app with no login must only be reachable from other machines on purpose.
func TestListenAddress(t *testing.T) {
	cases := []struct {
		host      string
		valid     bool
		address   string
		reachable bool
	}{
		{defaultHost, true, "127.0.0.1:8000", false},
		// The zero value must not fall open to every interface.
		{"", true, "127.0.0.1:8000", false},
		{"localhost", true, "localhost:8000", false},
		{"::1", true, "[::1]:8000", false},
		{"0.0.0.0", true, "0.0.0.0:8000", true},
		{"::", true, "[::]:8000", true},
		{"192.168.1.20", true, "192.168.1.20:8000", true},
		{"my-mac.local", false, "", false},
		{"127.0.0.1:9000", false, "", false},
	}

	for _, c := range cases {
		t.Run(c.host, func(t *testing.T) {
			// Load only validates a host that was set, so "" skips this.
			if got := validHost(c.host); c.host != "" && got != c.valid {
				t.Fatalf("validHost(%q) = %v, want %v", c.host, got, c.valid)
			}
			if !c.valid {
				return
			}
			conf := Config{Host: c.host, Port: 8000}
			if got := conf.Address(); got != c.address {
				t.Errorf("Address() = %q, want %q", got, c.address)
			}
			if got := conf.NetworkReachable(); got != c.reachable {
				t.Errorf("NetworkReachable() = %v, want %v", got, c.reachable)
			}
		})
	}
}

// TestSyncBackupsDir pins where snapshots land — the one place a database file
// is allowed inside the sync directory, because those are engine snapshots
// rather than copies of a file being appended to.
func TestSyncBackupsDir(t *testing.T) {
	if got := (Config{}).SyncBackupsDir(); got != "" {
		t.Errorf("with no sync dir, SyncBackupsDir() = %q, want empty", got)
	}
	want := filepath.Join("/tmp/sync", "backups")
	if got := (Config{SyncDir: "/tmp/sync"}).SyncBackupsDir(); got != want {
		t.Errorf("SyncBackupsDir() = %q, want %q", got, want)
	}
}
