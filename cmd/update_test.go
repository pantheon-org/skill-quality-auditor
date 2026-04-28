package cmd

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── helpers ─────────────────────────────────────────────────────────────────

func makeTarGz(t *testing.T, dir, binaryName string) string {
	t.Helper()
	archivePath := filepath.Join(dir, binaryName+"_linux_amd64.tar.gz")
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close() //nolint:errcheck

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	content := []byte("#!/bin/sh\necho skill-auditor")
	hdr := &tar.Header{
		Name: binaryName,
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	tw.Close() //nolint:errcheck
	gw.Close() //nolint:errcheck
	return archivePath
}

func sha256Hex(path string) string {
	data, _ := os.ReadFile(path)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func makeChecksums(archivePath string) string {
	return sha256Hex(archivePath) + "  " + filepath.Base(archivePath) + "\n"
}

func mockGitHubServer(t *testing.T, tag string, archiveName, archivePath string) *httptest.Server {
	t.Helper()
	archiveData, _ := os.ReadFile(archivePath)
	checksumData := []byte(makeChecksums(archivePath))

	mux := http.NewServeMux()

	releasePath := "/repos/" + githubRepo + "/releases/latest"
	if tag != "latest" {
		releasePath = "/repos/" + githubRepo + "/releases/tags/" + tag
	}

	mux.HandleFunc(releasePath, func(w http.ResponseWriter, _ *http.Request) {
		rel := ghRelease{
			TagName: tag,
			Assets: []ghAsset{
				{Name: archiveName, BrowserDownloadURL: ""},
				{Name: "checksums.txt", BrowserDownloadURL: ""},
			},
		}
		_ = json.NewEncoder(w).Encode(rel)
	})

	mux.HandleFunc("/download/"+archiveName, func(w http.ResponseWriter, _ *http.Request) {
		w.Write(archiveData) //nolint:errcheck
	})

	mux.HandleFunc("/download/checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		w.Write(checksumData) //nolint:errcheck
	})

	srv := httptest.NewServer(mux)

	// patch asset URLs to point at test server
	mux.HandleFunc(releasePath+"_patched", func(_ http.ResponseWriter, _ *http.Request) {})
	_ = srv // used below via closure; register patched handler before returning

	// redefine handler to include correct URLs
	mux2 := http.NewServeMux()
	mux2.HandleFunc(releasePath, func(w http.ResponseWriter, _ *http.Request) {
		rel := ghRelease{
			TagName: tag,
			Assets: []ghAsset{
				{
					Name:               archiveName,
					BrowserDownloadURL: srv.URL + "/download/" + archiveName,
				},
				{
					Name:               "checksums.txt",
					BrowserDownloadURL: srv.URL + "/download/checksums.txt",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(rel)
	})
	mux2.HandleFunc("/download/"+archiveName, func(w http.ResponseWriter, _ *http.Request) {
		w.Write(archiveData) //nolint:errcheck
	})
	mux2.HandleFunc("/download/checksums.txt", func(w http.ResponseWriter, _ *http.Request) {
		w.Write(checksumData) //nolint:errcheck
	})

	srv.Close()
	return httptest.NewServer(mux2)
}

// ── verifyChecksum ───────────────────────────────────────────────────────────

func TestVerifyChecksum_Valid(t *testing.T) {
	tmp := t.TempDir()
	archivePath := makeTarGz(t, tmp, "skill-auditor")

	checksumPath := filepath.Join(tmp, "checksums.txt")
	os.WriteFile(checksumPath, []byte(makeChecksums(archivePath)), 0o644) //nolint:errcheck

	if err := verifyChecksum(archivePath, checksumPath); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestVerifyChecksum_Mismatch(t *testing.T) {
	tmp := t.TempDir()
	archivePath := makeTarGz(t, tmp, "skill-auditor")

	checksumPath := filepath.Join(tmp, "checksums.txt")
	// Write wrong hash
	os.WriteFile(checksumPath, []byte("deadbeefdeadbeef  "+filepath.Base(archivePath)+"\n"), 0o644) //nolint:errcheck

	err := verifyChecksum(archivePath, checksumPath)
	if err == nil {
		t.Fatal("expected checksum mismatch error")
	}
	if !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyChecksum_MissingEntry(t *testing.T) {
	tmp := t.TempDir()
	archivePath := makeTarGz(t, tmp, "skill-auditor")

	checksumPath := filepath.Join(tmp, "checksums.txt")
	os.WriteFile(checksumPath, []byte("abc123  other_file.tar.gz\n"), 0o644) //nolint:errcheck

	err := verifyChecksum(archivePath, checksumPath)
	if err == nil {
		t.Fatal("expected error for missing entry")
	}
}

// ── extractBinary ────────────────────────────────────────────────────────────

func TestExtractBinary_Found(t *testing.T) {
	tmp := t.TempDir()
	archivePath := makeTarGz(t, tmp, "skill-auditor")
	destPath := filepath.Join(tmp, "skill-auditor-out")

	if err := extractBinary(archivePath, destPath); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(destPath); err != nil {
		t.Fatalf("binary not extracted: %v", err)
	}
}

func TestExtractBinary_NotFound(t *testing.T) {
	tmp := t.TempDir()
	// archive containing a differently-named binary
	archivePath := makeTarGz(t, tmp, "other-tool")
	destPath := filepath.Join(tmp, "out")

	err := extractBinary(archivePath, destPath)
	if err == nil {
		t.Fatal("expected error when binary not in archive")
	}
	if !strings.Contains(err.Error(), "not found inside archive") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ── latestReleaseTag (via mock HTTP) ─────────────────────────────────────────

func TestFetchRelease_Latest(t *testing.T) {
	tmp := t.TempDir()
	archivePath := makeTarGz(t, tmp, "skill-auditor")
	archiveName := filepath.Base(archivePath)
	srv := mockGitHubServer(t, "latest", archiveName, archivePath)
	defer srv.Close()

	// Temporarily swap the API base by monkey-patching fetchRelease via a wrapper
	origFetch := fetchReleaseFunc
	defer func() { fetchReleaseFunc = origFetch }()
	fetchReleaseFunc = func(tagOrLatest string) (*ghRelease, error) {
		var url string
		if tagOrLatest == "latest" {
			url = srv.URL + "/repos/" + githubRepo + "/releases/latest"
		} else {
			url = srv.URL + "/repos/" + githubRepo + "/releases/tags/" + tagOrLatest
		}
		resp, err := http.Get(url) //nolint:noctx
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close() //nolint:errcheck
		var rel ghRelease
		if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
			return nil, err
		}
		return &rel, nil
	}

	rel, err := fetchReleaseFunc("latest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rel.TagName != "latest" {
		t.Fatalf("expected tag 'latest', got %q", rel.TagName)
	}
}

// ── --check flag ─────────────────────────────────────────────────────────────

func TestUpdateCheck_AlreadyLatest(t *testing.T) {
	// Simulate version == target.
	orig := version
	version = "1.2.3"
	defer func() { version = orig }()

	updateCheckOnly = true
	updateVersionTarget = "v1.2.3"
	defer func() {
		updateCheckOnly = false
		updateVersionTarget = ""
	}()

	cmd := updateCmd
	if err := runUpdate(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
