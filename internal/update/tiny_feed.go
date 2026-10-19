package update

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

// Tiny consumes only its own GitHub releases. GitHub's asset digest binds
// the bytes downloaded by the inherited updater; absent digests fail closed.
const Site = "https://github.com/caelis-labs/tiny-magpie"
const releasesAPI = "https://api.github.com/repos/caelis-labs/tiny-magpie/releases?per_page=100"

func Feed() string {
	// Fixtures may use a loopback server. The upstream MAGPIE_UPDATE_FEED
	// override is deliberately ignored, including in an inherited environment.
	if f := os.Getenv("TINY_MAGPIE_UPDATE_FEED"); f != "" {
		if u, err := url.Parse(f); err == nil && u.Scheme == "http" && u.User == nil && net.ParseIP(u.Hostname()).IsLoopback() {
			return f
		}
	}
	return releasesAPI
}

type githubRelease struct {
	Tag     string `json:"tag_name"`
	URL     string `json:"html_url"`
	Body    string `json:"body"`
	Draft   bool   `json:"draft"`
	Preview bool   `json:"prerelease"`
	Assets  []struct {
		Name   string `json:"name"`
		URL    string `json:"browser_download_url"`
		Digest string `json:"digest"`
		Size   int64  `json:"size"`
	} `json:"assets"`
}

func releaseList(ctx context.Context) ([]githubRelease, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", Feed(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tiny-magpie releases: %s", res.Status)
	}
	var releases []githubRelease
	if err := json.NewDecoder(io.LimitReader(res.Body, 8<<20)).Decode(&releases); err != nil {
		return nil, fmt.Errorf("tiny-magpie releases: %w", err)
	}
	return releases, nil
}

var tinyTagVersion = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$`)

func tinyVersion(r githubRelease) string {
	v, ok := strings.CutPrefix(r.Tag, "tiny-v")
	if !ok || !tinyTagVersion.MatchString(v) || !Released(v) || r.Draft || r.URL != Site+"/releases/tag/"+r.Tag {
		return ""
	}
	return v
}

func LatestIn(ctx context.Context, lang string) (*Release, error) {
	return LatestInChannel(ctx, lang, "stable")
}

// Preview includes stable releases too. The default stable channel never
// takes a preview; only an explicit preview choice includes one.
func LatestInChannel(ctx context.Context, lang, channel string) (*Release, error) {
	list, err := releaseList(ctx)
	if err != nil {
		return nil, err
	}
	var best *githubRelease
	version := "0.0.0"
	for i := range list {
		r := &list[i]
		v := tinyVersion(*r)
		if v == "" || (channel != "preview" && (r.Preview || parse(v).pre != "")) {
			continue
		}
		if best == nil || Newer(v, version) {
			best, version = r, v
		}
	}
	if best == nil {
		return &Release{Version: version, URL: Site + "/releases"}, nil
	}
	out := &Release{Version: version, Notes: InLang(best.Body, lang), URL: best.URL, Assets: map[string]Asset{}}
	// Map versioned public filenames to the updater's logical platform names.
	// Nothing from another repo, tag, architecture or packaging format enters it.
	for _, name := range []string{AppAsset(), BinaryAsset()} {
		prefix := "tiny-magpie-"
		if strings.HasPrefix(name, "tiny-magpie-cli-") {
			prefix = "tiny-magpie-cli-"
		}
		actual := strings.Replace(name, prefix, prefix+version+"-", 1)
		for _, a := range best.Assets {
			if a.Name != actual {
				continue
			}
			if a.URL != Site+"/releases/download/"+best.Tag+"/"+actual || a.Size <= 0 {
				return nil, fmt.Errorf("invalid tiny-magpie asset: %s", actual)
			}
			hash, ok := strings.CutPrefix(a.Digest, "sha256:")
			if b, err := hex.DecodeString(hash); !ok || err != nil || len(b) != 32 {
				return nil, fmt.Errorf("tiny-magpie asset has no SHA-256 digest: %s", actual)
			}
			if _, exists := out.Assets[name]; exists {
				return nil, fmt.Errorf("duplicate tiny-magpie asset: %s", actual)
			}
			out.Assets[name] = Asset{URL: a.URL, Size: a.Size, SHA256: hash}
		}
	}
	return out, nil
}
