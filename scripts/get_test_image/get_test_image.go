package main

import (
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func main() {
	downloadTestImage(defaultTestImageSource, "testimage/full.qcow")
}

const defaultTestImageSource = "https://github.com/oVirt/ovirt-tinycore-linux"

//go:embed github.pem
var gitHubCerts []byte
var imageSourceGitHubPattern = regexp.MustCompile(`^https://github.com/([^/]+)/([^/]+)/?$`)
var testImageDownloadHTTPClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return nil
	},
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    getTestImageDownloadCertPool(),
			MinVersion: tls.VersionTLS12,
		},
	},
}

func getTestImageDownloadCertPool() *x509.CertPool {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		// This will happen on Windows where system pools are not supported before Go 1.18.
		certPool = x509.NewCertPool()
		certPool.AppendCertsFromPEM(gitHubCerts)
	}
	return certPool
}

func downloadTestImage(source string, target string) {
	_, err := os.Stat(target)
	if err == nil {
		// Already exists, no need to download
		log.Printf("Test image already exists at %s, skipping download.", target)
		return
	}
	log.Printf("Downloading test image from %s to %s...", source, target)
	// There is no file inclusion vulnerability here.
	targetFh, err := os.Create(target) // nolint:gosec
	if err != nil {
		log.Fatalf("failed to create temporary image file at %s (%v)", target, err)
	}

	if imageSourceGitHubPattern.MatchString(source) {
		source = getGitHubReleaseFileName(source)
		log.Printf("Downloading from %s...", source)
	}

	req, err := http.NewRequest("GET", source, nil)
	if err != nil {
		log.Fatalf("failed to create HTTP request for %s (%v)", source, err)
	}
	if ghToken := os.Getenv("GITHUB_TOKEN"); ghToken != "" {
		req.Header.Add("authorization", fmt.Sprintf("bearer %s", ghToken))
	}
	resp, err := testImageDownloadHTTPClient.Do(req)
	if err != nil {
		log.Fatalf("failed to download test image %s (%v)", source, err)
	}
	if resp.StatusCode != 200 {
		log.Fatalf("failed to download test image %s (status code is %d)", source, err)
	}

	if _, err := io.Copy(targetFh, resp.Body); err != nil {
		log.Fatalf("failed to copy from %s to %s (%v)", source, target, err)
	}
	_ = resp.Body.Close()
	if err := targetFh.Close(); err != nil {
		log.Fatalf("failed to close temporary image file at %s (%v)", target, err)
	}
	log.Printf("Download complete.")
}

func getGitHubReleaseFileName(source string) string {
	log.Printf("Getting latest release from GitHub repo %s...", source)
	matches := imageSourceGitHubPattern.FindStringSubmatch(source)
	if len(matches) == 2 {
		log.Fatalf("invalid GitHub source URL: %s", source)
	}
	org := matches[1]
	repo := matches[2]
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", org, repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("failed to query releases for %s (%v)", source, err)
	}
	if ghToken := os.Getenv("GITHUB_TOKEN"); ghToken != "" {
		req.Header.Add("authorization", fmt.Sprintf("bearer %s", ghToken))
	}
	resp, err := testImageDownloadHTTPClient.Do(req)
	if err != nil {
		log.Fatalf("failed to query releases for %s (%v)", source, err)
	}
	if resp.StatusCode != 200 {
		_ = resp.Body.Close()
		log.Fatalf(
			"GitHub responded with a non-200 status code on the latest release API for %s (%d)",
			source,
			resp.StatusCode,
		)
	}
	jsonReader := json.NewDecoder(resp.Body)
	var release gitHubRelease
	if err := jsonReader.Decode(&release); err != nil {
		_ = resp.Body.Close()
		log.Fatalf(
			"GitHub responded with an invalid JSON for the latest release %s (%v)",
			source,
			resp.StatusCode,
		)
	}

	for _, asset := range release.Assets {
		if strings.HasSuffix(asset.Name, ".qcow") || strings.HasSuffix(asset.Name, ".qcow2") {
			_ = resp.Body.Close()
			return asset.BrowserDownloadURL
		}
	}
	_ = resp.Body.Close()
	log.Fatalf("No QCOW asset found on the latest release for %s.", source)
	return ""
}

type gitHubRelease struct {
	AssetsURL string        `json:"assets_url"`
	TagName   string        `json:"tag_name"`
	Name      string        `json:"name"`
	Assets    []gitHubAsset `json:"assets"`
}

type gitHubAsset struct {
	BrowserDownloadURL string `json:"browser_download_url"`
	Name               string `json:"name"`
	ContentType        string `json:"content_type"`
	Size               uint   `json:"size"`
}
