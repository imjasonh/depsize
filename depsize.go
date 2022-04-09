package depsize

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

func proxy() string {
	proxy := os.Getenv("GOPROXY")
	if proxy != "" {
		if idx := strings.IndexRune(proxy, ','); idx > 0 {
			proxy = proxy[:idx]
		}
		return proxy
	}
	return "https://proxy.golang.org"
}

// Size returns the size of the dep module, at the given version, in bytes.
func Size(dep, version string) (int, error) {
	dep = strings.ToLower(dep)
	proxy := proxy()

	if version == "latest" {
		var err error
		version, err = Latest(dep)
		if err != nil {
			return 0, err
		}
	}
	url := fmt.Sprintf("%s/%s/@v/%s.zip", proxy, dep, version)
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("HEAD %s: %d %s", url, resp.StatusCode, resp.Status)
	}
	l := resp.Header.Get("Content-Length")
	if l == "" {
		return 0, errors.New("missing content-length header")
	}
	return strconv.Atoi(l)
}

func Latest(dep string) (string, error) {
	proxy := proxy()
	url := fmt.Sprintf("%s/%s/@latest", proxy, dep)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GET %s: %d %s", url, resp.StatusCode, resp.Status)
	}
	defer resp.Body.Close()
	var info struct{ Version string }
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("decoding JSON: %w", err)
	}
	return info.Version, nil
}

type Dep struct {
	Dep, Version string
	Size         int
}

// Deps returns the transitive deps of a module.
func Deps(dep, version string) ([]Dep, error) {
	proxy := proxy()
	if version == "latest" {
		var err error
		version, err = Latest(dep)
		if err != nil {
			return nil, err
		}
	}

	url := fmt.Sprintf("%s/%s/@v/%s.mod", proxy, dep, version)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %d %s", url, resp.StatusCode, resp.Status)
	}
	defer resp.Body.Close()
	all, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	mod, err := modfile.Parse("", all, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing go.mod file: %w", err)
	}

	// Handle replaces, sort and return.
	m := map[module.Version]struct{}{}
	for _, r := range mod.Require {
		m[r.Mod] = struct{}{}
	}
	for _, r := range mod.Replace {
		delete(m, r.Old)
		if strings.HasPrefix(r.New.Path, "../") {
			continue // Ignore local replaces
		}
		m[r.New] = struct{}{}
	}
	var out []Dep
	for k := range m {
		size, err := Size(k.Path, k.Version)
		if err != nil {
			return nil, fmt.Errorf("getting size for %s %s: %w", k.Path, k.Version, err)
		}
		out = append(out, Dep{
			Dep:     k.Path,
			Version: k.Version,
			Size:    size,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Size > out[j].Size
	})
	return out, nil
}
