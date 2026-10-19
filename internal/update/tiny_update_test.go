package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yetone/magpie/internal/proc"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func tinyFixture(version string, preview bool) map[string]any {
	tag := "tiny-v" + version
	names := []string{AppAsset(), BinaryAsset()}
	assets := []map[string]any{}
	for _, name := range names {
		prefix := "tiny-magpie-"
		if strings.HasPrefix(name, "tiny-magpie-cli-") {
			prefix = "tiny-magpie-cli-"
		}
		actual := strings.Replace(name, prefix, prefix+version+"-", 1)
		assets = append(assets, map[string]any{"name": actual, "browser_download_url": Site + "/releases/download/" + tag + "/" + actual, "digest": "sha256:" + strings.Repeat("a", 64), "size": 123})
	}
	return map[string]any{"tag_name": tag, "html_url": ReleasePage(version), "prerelease": preview, "body": "English\n<!-- lang:zh -->\n中文", "assets": assets}
}

func tinyServer(t *testing.T, value any) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { json.NewEncoder(w).Encode(value) }))
	t.Cleanup(srv.Close)
	t.Setenv("TINY_MAGPIE_UPDATE_FEED", srv.URL)
}

func TestTinyFeedBoundary(t *testing.T) {
	t.Setenv("MAGPIE_UPDATE_FEED", "https://usemagpie.ai/api/latest")
	t.Setenv("MAGPIE_NOTES_FEED", "https://usemagpie.ai/api/notes")
	for _, bad := range []string{"https://usemagpie.ai/api/latest", "http://localhost.evil.test/", "file:///tmp/feed", "http://user@127.0.0.1/", ""} {
		t.Setenv("TINY_MAGPIE_UPDATE_FEED", bad)
		if Feed() != releasesAPI || NotesFeed() != releasesAPI {
			t.Fatal("escaped tiny release feed", Feed(), NotesFeed())
		}
	}
	if ReleasePage("0.1.2") != Site+"/releases/tag/tiny-v0.1.2" {
		t.Fatal("wrong release page")
	}
}

func TestTinyReleaseChannelsAndAssets(t *testing.T) {
	foreign := tinyFixture("99.0.0", false)
	foreign["html_url"] = "https://github.com/yetone/magpie/releases/tag/v99.0.0"
	draft := tinyFixture("98.0.0", false)
	draft["draft"] = true
	tinyServer(t, []any{tinyFixture("1.9.0", false), foreign, draft, tinyFixture("1.10.0", true), tinyFixture("2.0.0-rc.1", false)})
	for channel, want := range map[string]string{"": "1.9.0", "stable": "1.9.0", "preview": "2.0.0-rc.1"} {
		r, err := LatestInChannel(context.Background(), "zh", channel)
		if err != nil {
			t.Fatal(err)
		}
		if r.Version != want || r.Notes != "中文" {
			t.Fatalf("%s: %+v", channel, r)
		}
		for _, name := range []string{AppAsset(), BinaryAsset()} {
			a, ok := r.Assets[name]
			if !ok || !strings.Contains(a.URL, want) || a.SHA256 != strings.Repeat("a", 64) {
				t.Fatalf("wrong platform artifact: %s %+v", name, a)
			}
		}
	}
	notes, err := NotesBetween(context.Background(), "1.9.0", "1.10.0", "zh")
	if err != nil || len(notes) != 1 || notes[0].Version != "1.10.0" {
		t.Fatalf("notes: %+v %v", notes, err)
	}
}

func TestTinyReleaseRejectsUntrustedAsset(t *testing.T) {
	for _, kind := range []string{"foreign-url", "wrong-tag", "no-digest", "bad-digest", "empty-size", "duplicate"} {
		t.Run(kind, func(t *testing.T) {
			r := tinyFixture("1.0.0", false)
			assets := r["assets"].([]map[string]any)
			a := assets[0]
			switch kind {
			case "foreign-url":
				a["browser_download_url"] = "https://github.com/yetone/magpie/releases/download/v1.0.0/a.dmg"
			case "wrong-tag":
				a["browser_download_url"] = strings.Replace(a["browser_download_url"].(string), "tiny-v1.0.0", "tiny-v0.9.0", 1)
			case "no-digest":
				delete(a, "digest")
			case "bad-digest":
				a["digest"] = "sha256:garbage"
			case "empty-size":
				a["size"] = 0
			case "duplicate":
				assets = append(assets, a)
				r["assets"] = assets
			}
			tinyServer(t, []any{r})
			if _, err := Latest(context.Background()); err == nil {
				t.Fatal("accepted untrusted asset", kind)
			}
		})
	}
}

func TestTinyStableWithoutRelease(t *testing.T) {
	tinyServer(t, []any{tinyFixture("1.0.0", true)})
	r, err := LatestInChannel(context.Background(), "", "stable")
	if err != nil || r.Version != "0.0.0" || len(r.Assets) != 0 {
		t.Fatalf("unexpected stable release %+v %v", r, err)
	}
}

func TestTinyDownloadTamperDoesNotLeaveCandidate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "tampered") }))
	defer srv.Close()
	hash := sha256.Sum256([]byte("original"))
	dest := filepath.Join(t.TempDir(), "update.dmg")
	err := fetch(context.Background(), Asset{URL: srv.URL, SHA256: hex.EncodeToString(hash[:])}, dest)
	if err == nil {
		t.Fatal("tampered update accepted")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatal("tampered candidate retained", err)
	}
}

func TestTinySignerAndBundleIdentity(t *testing.T) {
	old := signatureCommand
	t.Cleanup(func() { signatureCommand = old })
	for _, bad := range []string{"", "unsigned", "other-team", "original-magpie", "bad-signature", "gatekeeper", "wrong-version"} {
		t.Run(bad, func(t *testing.T) {
			signatureCommand = func(ctx context.Context, name string, args ...string) ([]byte, error) {
				switch name {
				case "codesign":
					if args[0] == "-dv" {
						team := "64KZ67PM5J"
						if bad == "unsigned" {
							team = ""
						}
						if bad == "other-team" {
							team = "WRONG"
						}
						return []byte("TeamIdentifier=" + team), nil
					}
					if bad == "bad-signature" {
						return nil, errors.New("invalid signature")
					}
				case "defaults":
					if args[len(args)-1] == "CFBundleShortVersionString" {
						if bad == "wrong-version" {
							return []byte("2.0.0"), nil
						}
						return []byte("1.0.0"), nil
					}
					if bad == "original-magpie" {
						return []byte("ai.magpie.app"), nil
					}
					return []byte("dev.caelis.tiny-magpie"), nil
				case "spctl":
					if bad == "gatekeeper" {
						return nil, errors.New("rejected")
					}
				}
				return nil, nil
			}
			err := verifyApp(context.Background(), "installed.app", "candidate.app", "1.0.0")
			if (err != nil) != (bad != "") {
				t.Fatalf("trust case %s: %v", bad, err)
			}
		})
	}
}

// Opt-in local artifact acceptance; no network other than its loopback server,
// no credentials, and only temporary copies are ever installed. CI skips this
// when the previously verified signed fixture paths are absent.
func TestTinySignedDMGFixture(t *testing.T) {
	dmg, sourceApp := os.Getenv("TINY_SIGNED_DMG_FIXTURE"), os.Getenv("TINY_SIGNED_APP_FIXTURE")
	if dmg == "" || sourceApp == "" || runtime.GOOS != "darwin" {
		t.Skip("requires explicitly supplied local signed fixtures")
	}
	root := t.TempDir()
	bundle := filepath.Join(root, "installed", "tiny-magpie.app")
	if b, err := proc.Command("ditto", sourceApp, bundle).CombinedOutput(); err != nil {
		t.Fatalf("copy fixture: %v %s", err, b)
	}
	file, err := os.Open(dmg)
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	size, err := io.Copy(h, file)
	file.Close()
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, dmg) }))
	defer srv.Close()
	rel := &Release{Version: "0.1.2", Assets: map[string]Asset{AppAsset(): {URL: srv.URL, Size: size, SHA256: hex.EncodeToString(h.Sum(nil))}}}
	staged, err := Stage(context.Background(), rel, bundle)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(staged, ".tiny-magpie-update") {
		t.Fatal("shared upstream staging path", staged)
	}
	if err := Install(staged, bundle); err != nil {
		t.Fatal(err)
	}
	if err := verifyApp(context.Background(), bundle, bundle, "0.1.2"); err != nil {
		t.Fatal(err)
	}
	t.Log("signed DMG hash, mount, team, bundle ID, version, Gatekeeper and temporary-copy install passed")
}

func TestTinyAdminPromptIdentity(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS administrator dialog")
	}
	previous := run
	t.Cleanup(func() { run = previous })
	var command string
	run = func(name string, args ...string) ([]byte, error) {
		if name != "osascript" || len(args) != 2 || args[0] != "-e" {
			t.Fatalf("unexpected administrator launcher: %s %v", name, args)
		}
		command = args[1]
		return nil, nil
	}
	if err := asAdmin(`mv '/tmp/staged.app' '/Applications/tiny-magpie.app'`); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(command, `with prompt "tiny-magpie is updating itself."`) {
		t.Fatalf("administrator prompt uses another product's name: %s", command)
	}
}
