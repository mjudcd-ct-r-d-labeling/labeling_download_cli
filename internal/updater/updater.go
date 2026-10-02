// Package updater installs verified, publicly published CLI releases.
package updater

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/mjudcd-ct-r-d-labeling/labeling_download_cli/internal/version"
)

const (
	tagsURL       = "https://api.github.com/repos/mjudcd-ct-r-d-labeling/labeling_download_cli/tags"
	releaseBase   = "https://mjudcd-grac-api.newlearn.ai.kr"
	maxBinarySize = 256 << 20
)

type release struct {
	URL      string `json:"download_url"`
	Filename string `json:"file_name"`
	Size     int64  `json:"file_size"`
	SHA256   string `json:"sha256"`
}

// Run performs an explicit update, independently of dataset authentication.
func Run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.ResponseHeaderTimeout = 30 * time.Second
	c := &http.Client{Transport: t}
	fmt.Println("Checking for CLI updates...")
	latest, err := latestVersion(ctx, c, tagsURL)
	if err != nil {
		return err
	}
	if version.String() != "dev" {
		comparison, err := compareVersions(latest, version.String())
		if err != nil {
			return fmt.Errorf("cannot compare installed version; reinstall using the official installer")
		}
		if comparison <= 0 {
			fmt.Printf("Already up to date (%s).\n", version.String())
			return nil
		}
	}
	platform, filename, err := platformBinary(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	var info release
	if err := getJSON(ctx, c, releaseBase+"/cli-releases/download/"+platform+"/"+latest, &info); err != nil {
		return err
	}
	if info.Filename != filename {
		return fmt.Errorf("release does not match this platform")
	}
	target, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot locate installed CLI")
	}
	target, err = filepath.EvalSymlinks(target)
	if err != nil {
		return fmt.Errorf("cannot resolve installed CLI path")
	}
	fmt.Printf("Downloading mju-dataset %s...\n", latest)
	source, err := download(ctx, c, info)
	if err != nil {
		return err
	}
	defer os.Remove(source)
	if err := ctx.Err(); err != nil {
		return err
	}
	return install(ctx, source, target, latest)
}

func platformBinary(goos, arch string) (string, string, error) {
	platform := goos + "-" + arch
	switch platform {
	case "darwin-amd64", "darwin-arm64", "linux-amd64", "linux-arm64", "windows-amd64":
	default:
		return "", "", fmt.Errorf("CLI updates are not available for this platform")
	}
	name := "mju-dataset-" + platform
	if goos == "windows" {
		name += ".exe"
	}
	return platform, name, nil
}

func request(ctx context.Context, c *http.Client, address string) (*http.Response, error) {
	u, err := url.Parse(address)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return nil, fmt.Errorf("invalid release URL")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid release request")
	}
	req.Header.Set("User-Agent", "mju-dataset-updater")
	resp, err := c.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("update network error; please try again")
	}
	if resp.Request.URL.Scheme != "https" || resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("release request failed (HTTP %d)", resp.StatusCode)
	}
	return resp, nil
}

func getJSON(ctx context.Context, c *http.Client, address string, out any) error {
	resp, err := request(ctx, c, address)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(out); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("invalid release metadata")
	}
	return nil
}

func download(ctx context.Context, c *http.Client, info release) (path string, err error) {
	expected, err := hex.DecodeString(info.SHA256)
	if err != nil || len(expected) != sha256.Size || info.Size <= 0 || info.Size > maxBinarySize {
		return "", fmt.Errorf("release checksum or size is missing or invalid")
	}
	resp, err := request(ctx, c, info.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	f, err := os.CreateTemp("", "mju-dataset-update-*")
	if err != nil {
		return "", fmt.Errorf("cannot create update temporary file")
	}
	path = f.Name()
	defer func() {
		f.Close()
		if err != nil {
			os.Remove(path)
		}
	}()
	hash := sha256.New()
	n, err := io.Copy(io.MultiWriter(f, hash), io.LimitReader(resp.Body, info.Size+1))
	if err != nil {
		if ctx.Err() != nil {
			return path, ctx.Err()
		}
		return path, fmt.Errorf("update download failed")
	}
	if n != info.Size || !equalHash(hash.Sum(nil), expected) {
		return path, fmt.Errorf("update integrity verification failed; installed CLI was preserved")
	}
	if err := f.Sync(); err != nil {
		return path, fmt.Errorf("cannot save update")
	}
	if err := f.Close(); err != nil {
		return path, fmt.Errorf("cannot close update file")
	}
	return path, nil
}

func equalHash(a, b []byte) bool {
	return hex.EncodeToString(a) == hex.EncodeToString(b)
}

// stage copies into the installation directory so the final rename is atomic.
func stage(source, target string) (path string, err error) {
	in, err := os.Open(source)
	if err != nil {
		return "", err
	}
	defer in.Close()
	out, err := os.CreateTemp(filepath.Dir(target), ".mju-dataset-update-*")
	if err != nil {
		return "", err
	}
	path = out.Name()
	defer func() {
		out.Close()
		if err != nil {
			os.Remove(path)
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return path, err
	}
	if err = out.Chmod(0755); err != nil {
		return path, err
	}
	if err = out.Sync(); err != nil {
		return path, err
	}
	err = out.Close()
	return path, err
}
