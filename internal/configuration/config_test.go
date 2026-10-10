package configuration

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestGet_ReloadsEditedFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "system-control.yaml")
	write := func(content string, mtime time.Time) {
		t.Helper()
		if err := os.WriteFile(file, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(file, mtime, mtime); err != nil {
			t.Fatal(err)
		}
		// let Get look at the file right away
		lastChecked = time.Time{}
	}

	viper.Reset()
	t.Cleanup(func() { viper.Reset(); CurrentConfig = Configuration{}; fileStamp = fileVersion{} })

	base := time.Now().Add(-time.Hour)
	write("fan:\n  policyFileGlob: /one\n", base)
	InitConfig(file)
	if err := Reload(); err != nil {
		t.Fatal(err)
	}
	if got := Get().Fan.PolicyFileGlob; got != "/one" {
		t.Fatalf("glob = %q", got)
	}

	write("fan:\n  policyFileGlob: /two\n", base.Add(time.Minute))
	if got := Get().Fan.PolicyFileGlob; got != "/two" {
		t.Errorf("edit not noticed, glob = %q", got)
	}

	// broken file keeps the previous values
	write("fan: [unclosed\n", base.Add(2*time.Minute))
	if got := Get().Fan.PolicyFileGlob; got != "/two" {
		t.Errorf("broken file replaced the configuration, glob = %q", got)
	}
}
