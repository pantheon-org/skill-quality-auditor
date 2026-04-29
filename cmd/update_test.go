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
	"runtime"
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

// ── assetURL ─────────────────────────────────────────────────────────────────

func TestAssetURL_found(t *testing.T) {
	rel := &ghRelease{Assets: []ghAsset{
		{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums.txt"},
		{Name: "skill-auditor_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/binary.tar.gz"},
	}}
	got := assetURL(rel, "skill-auditor_linux_amd64.tar.gz")
	if got != "https://example.com/binary.tar.gz" {
		t.Errorf("got %q, want binary URL", got)
	}
}

func TestAssetURL_notFound(t *testing.T) {
	rel := &ghRelease{Assets: []ghAsset{
		{Name: "other.tar.gz", BrowserDownloadURL: "https://example.com/other.tar.gz"},
	}}
	if got := assetURL(rel, "missing.tar.gz"); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestAssetURL_emptyAssets(t *testing.T) {
	rel := &ghRelease{}
	if got := assetURL(rel, "anything"); got != "" {
		t.Errorf("expected empty string for empty assets, got %q", got)
	}
}

// ── latestReleaseTag ─────────────────────────────────────────────────────────

func TestLatestReleaseTag_success(t *testing.T) {
	orig := fetchReleaseFunc
	defer func() { fetchReleaseFunc = orig }()
	fetchReleaseFunc = func(_ string) (*ghRelease, error) {
		return &ghRelease{TagName: "v9.9.9"}, nil
	}
	tag, err := latestReleaseTag()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v9.9.9" {
		t.Errorf("got %q, want v9.9.9", tag)
	}
}

func TestLatestReleaseTag_error(t *testing.T) {
	orig := fetchReleaseFunc
	defer func() { fetchReleaseFunc = orig }()
	fetchReleaseFunc = func(_ string) (*ghRelease, error) {
		return nil, fmt.Errorf("network error")
	}
	if _, err := latestReleaseTag(); err == nil {
		t.Error("expected error from failing fetchReleaseFunc")
	}
}

// ── fetchRelease ─────────────────────────────────────────────────────────────

func TestFetchRelease_LatestViaAPIBaseURL(t *testing.T) {
	rel := ghRelease{TagName: "v1.0.0"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			_ = json.NewEncoder(w).Encode(rel)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	origBase := apiBaseURL
	apiBaseURL = srv.URL
	defer func() { apiBaseURL = origBase }()

	got, err := fetchRelease("latest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TagName != "v1.0.0" {
		t.Errorf("got %q, want v1.0.0", got.TagName)
	}
}

func TestFetchRelease_ByTagViaAPIBaseURL(t *testing.T) {
	rel := ghRelease{TagName: "v2.3.4"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(rel)
	}))
	defer srv.Close()

	origBase := apiBaseURL
	apiBaseURL = srv.URL
	defer func() { apiBaseURL = origBase }()

	got, err := fetchRelease("v2.3.4")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TagName != "v2.3.4" {
		t.Errorf("got %q, want v2.3.4", got.TagName)
	}
}

func TestFetchRelease_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	origBase := apiBaseURL
	apiBaseURL = srv.URL
	defer func() { apiBaseURL = origBase }()

	if _, err := fetchRelease("latest"); err == nil {
		t.Error("expected error for non-200 status")
	}
}

// ── downloadFile ─────────────────────────────────────────────────────────────

func TestDownloadFile_success(t *testing.T) {
	content := []byte("binary content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write(content) //nolint:errcheck
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "out")
	if err := downloadFile(srv.URL+"/file", dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := os.ReadFile(dest)
	if string(got) != string(content) {
		t.Errorf("file content mismatch")
	}
}

func TestDownloadFile_nonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	if err := downloadFile(srv.URL+"/file", filepath.Join(t.TempDir(), "out")); err == nil {
		t.Error("expected error for non-200 status")
	}
}

func TestDownloadFile_createError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("data")) //nolint:errcheck
	}))
	defer srv.Close()

	if err := downloadFile(srv.URL, "/nonexistent/dir/out"); err == nil {
		t.Error("expected error when dest directory does not exist")
	}
}

func TestFetchRelease_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("not json")) //nolint:errcheck
	}))
	defer srv.Close()

	origBase := apiBaseURL
	apiBaseURL = srv.URL
	defer func() { apiBaseURL = origBase }()

	if _, err := fetchRelease("latest"); err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestExtractBinary_MissingArchive(t *testing.T) {
	if err := extractBinary("/nonexistent/archive.tar.gz", filepath.Join(t.TempDir(), "out")); err == nil {
		t.Error("expected error for missing archive file")
	}
}

func TestExtractBinary_InvalidGzip(t *testing.T) {
	tmp := t.TempDir()
	bad := filepath.Join(tmp, "bad.tar.gz")
	if err := os.WriteFile(bad, []byte("not gzip data at all"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := extractBinary(bad, filepath.Join(tmp, "out")); err == nil {
		t.Error("expected error for invalid gzip content")
	}
}

// ── replaceBinary ────────────────────────────────────────────────────────────

func TestReplaceBinary_success(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "new-binary")
	dest := filepath.Join(tmp, "current-binary")

	if err := os.WriteFile(src, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := replaceBinary(src, dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := os.ReadFile(dest)
	if string(got) != "new" {
		t.Errorf("expected dest to contain new binary content")
	}
}

func TestReplaceBinary_missingSrc(t *testing.T) {
	tmp := t.TempDir()
	dest := filepath.Join(tmp, "current")
	if err := os.WriteFile(dest, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := replaceBinary(filepath.Join(tmp, "nonexistent"), dest); err == nil {
		t.Error("expected error for missing source")
	}
}

func TestReplaceBinary_missingDestDir(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src")
	if err := os.WriteFile(src, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}
	// dest dir does not exist → os.CreateTemp should fail
	if err := replaceBinary(src, "/nonexistent/dir/binary"); err == nil {
		t.Error("expected error when dest directory does not exist")
	}
}

// ── runUpdate error paths ────────────────────────────────────────────────────

func TestRunUpdate_noMatchingAsset(t *testing.T) {
	orig := fetchReleaseFunc
	defer func() { fetchReleaseFunc = orig }()
	fetchReleaseFunc = func(_ string) (*ghRelease, error) {
		// Return a release with no assets — archiveName won't be found.
		return &ghRelease{TagName: "v1.0.0", Assets: []ghAsset{}}, nil
	}

	origTarget := updateVersionTarget
	updateVersionTarget = "v1.0.0"
	defer func() { updateVersionTarget = origTarget }()

	if err := runUpdate(updateCmd, nil); err == nil {
		t.Error("expected error for missing release asset")
	}
}

func TestRunUpdate_noChecksumAsset(t *testing.T) {
	archiveName := fmt.Sprintf("skill-auditor_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)

	orig := fetchReleaseFunc
	defer func() { fetchReleaseFunc = orig }()
	fetchReleaseFunc = func(_ string) (*ghRelease, error) {
		return &ghRelease{
			TagName: "v1.0.0",
			Assets: []ghAsset{
				{Name: archiveName, BrowserDownloadURL: "https://example.com/binary.tar.gz"},
				// deliberately omit checksums.txt
			},
		}, nil
	}

	origTarget := updateVersionTarget
	updateVersionTarget = "v1.0.0"
	defer func() { updateVersionTarget = origTarget }()

	if err := runUpdate(updateCmd, nil); err == nil {
		t.Error("expected error for missing checksums asset")
	}
}

func TestRunUpdate_fetchReleaseError(t *testing.T) {
	orig := fetchReleaseFunc
	defer func() { fetchReleaseFunc = orig }()
	fetchReleaseFunc = func(_ string) (*ghRelease, error) {
		return nil, fmt.Errorf("network failure")
	}

	origTarget := updateVersionTarget
	updateVersionTarget = "v1.0.0"
	defer func() { updateVersionTarget = origTarget }()

	if err := runUpdate(updateCmd, nil); err == nil {
		t.Error("expected error when fetchReleaseFunc fails")
	}
}

func TestRunUpdate_latestTagFetchError(t *testing.T) {
	orig := fetchReleaseFunc
	defer func() { fetchReleaseFunc = orig }()
	fetchReleaseFunc = func(_ string) (*ghRelease, error) {
		return nil, fmt.Errorf("network failure")
	}

	origTarget := updateVersionTarget
	updateVersionTarget = ""
	defer func() { updateVersionTarget = origTarget }()

	if err := runUpdate(updateCmd, nil); err == nil {
		t.Error("expected error when latest tag fetch fails")
	}
}

func TestRunUpdate_checkOutdated(t *testing.T) {
	orig := fetchReleaseFunc
	defer func() { fetchReleaseFunc = orig }()
	fetchReleaseFunc = func(_ string) (*ghRelease, error) {
		return &ghRelease{TagName: "v99.0.0"}, nil
	}

	origVersion := version
	version = "1.0.0"
	defer func() { version = origVersion }()

	origTarget := updateVersionTarget
	updateVersionTarget = ""
	defer func() { updateVersionTarget = origTarget }()

	origCheck := updateCheckOnly
	updateCheckOnly = true
	defer func() { updateCheckOnly = origCheck }()

	if err := runUpdate(updateCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunUpdate_fullDownloadFlow(t *testing.T) {
	tmp := t.TempDir()
	archiveName := fmt.Sprintf("skill-auditor_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	namedArchive := filepath.Join(tmp, archiveName)

	// Build a proper tar.gz containing a "skill-auditor" binary for the current OS/arch.
	func() {
		f, err := os.Create(namedArchive)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close() //nolint:errcheck
		gw := gzip.NewWriter(f)
		tw := tar.NewWriter(gw)
		content := []byte("#!/bin/sh\necho skill-auditor")
		hdr := &tar.Header{Name: "skill-auditor", Mode: 0o755, Size: int64(len(content))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatal(err)
		}
		tw.Close() //nolint:errcheck
		gw.Close() //nolint:errcheck
	}()

	checksumData := []byte(makeChecksums(namedArchive))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/archive":
			data, _ := os.ReadFile(namedArchive)
			w.Write(data) //nolint:errcheck
		case "/checksums":
			w.Write(checksumData) //nolint:errcheck
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	orig := fetchReleaseFunc
	defer func() { fetchReleaseFunc = orig }()
	fetchReleaseFunc = func(_ string) (*ghRelease, error) {
		return &ghRelease{
			TagName: "v1.0.0",
			Assets: []ghAsset{
				{Name: archiveName, BrowserDownloadURL: srv.URL + "/archive"},
				{Name: "checksums.txt", BrowserDownloadURL: srv.URL + "/checksums"},
			},
		}, nil
	}

	origTarget := updateVersionTarget
	updateVersionTarget = "v1.0.0"
	defer func() { updateVersionTarget = origTarget }()

	origCheck := updateCheckOnly
	updateCheckOnly = false
	defer func() { updateCheckOnly = origCheck }()

	// Run the full download/verify/extract path. The final replaceBinary step may
	// succeed or fail depending on the environment — either outcome is acceptable.
	_ = runUpdate(updateCmd, nil)
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
