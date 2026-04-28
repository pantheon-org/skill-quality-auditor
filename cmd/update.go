package cmd

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const githubRepo = "pantheon-org/skill-quality-auditor"

var (
	updateCheckOnly     bool
	updateVersionTarget string

	// fetchReleaseFunc is a variable so tests can substitute a mock.
	fetchReleaseFunc = fetchRelease
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update skill-auditor to the latest (or a specific) release",
	Long: `Fetch the latest release from GitHub and replace the running binary.

Only applicable when installed via the install.sh script.
Homebrew users: brew upgrade skill-auditor
mise users:     mise upgrade skill-auditor`,
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().BoolVar(&updateCheckOnly, "check", false, "report the latest version without installing")
	updateCmd.Flags().StringVar(&updateVersionTarget, "version-target", "", "install a specific version (e.g. v1.2.3)")
	rootCmd.AddCommand(updateCmd)
}

// ghRelease is the subset of the GitHub releases API we care about.
type ghRelease struct {
	TagName string    `json:"tag_name"`
	Assets  []ghAsset `json:"assets"`
}

type ghAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func runUpdate(_ *cobra.Command, _ []string) error {
	target := updateVersionTarget
	if target == "" {
		var err error
		target, err = latestReleaseTag()
		if err != nil {
			return fmt.Errorf("could not resolve latest release: %w", err)
		}
	}

	current := version
	if updateCheckOnly {
		if current == target || "v"+current == target {
			fmt.Printf("skill-auditor %s is already the latest\n", current)
		} else {
			fmt.Printf("skill-auditor %s → %s available\n", current, target)
		}
		return nil
	}

	if current == target || "v"+current == target {
		fmt.Printf("skill-auditor %s is already the latest\n", current)
		return nil
	}

	fmt.Printf("Updating skill-auditor %s → %s\n", current, target)

	release, err := fetchReleaseFunc(target)
	if err != nil {
		return fmt.Errorf("could not fetch release %s: %w", target, err)
	}

	archiveName := fmt.Sprintf("skill-auditor_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	checksumName := "checksums.txt"

	archiveURL := assetURL(release, archiveName)
	if archiveURL == "" {
		return fmt.Errorf("no release asset found for %s/%s (looked for %s)", runtime.GOOS, runtime.GOARCH, archiveName)
	}
	checksumURL := assetURL(release, checksumName)
	if checksumURL == "" {
		return fmt.Errorf("checksums.txt not found in release %s", target)
	}

	tmp, err := os.MkdirTemp("", "skill-auditor-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp) //nolint:errcheck // cleanup, not critical

	archivePath := filepath.Join(tmp, archiveName)
	checksumPath := filepath.Join(tmp, checksumName)

	fmt.Printf("  Downloading %s...\n", archiveName)
	if err := downloadFile(archiveURL, archivePath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	if err := downloadFile(checksumURL, checksumPath); err != nil {
		return fmt.Errorf("download checksums failed: %w", err)
	}

	fmt.Println("  Verifying checksum...")
	if err := verifyChecksum(archivePath, checksumPath); err != nil {
		return err
	}

	binaryPath := filepath.Join(tmp, "skill-auditor")
	if err := extractBinary(archivePath, binaryPath); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	selfPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine current executable path: %w", err)
	}
	selfPath, err = filepath.EvalSymlinks(selfPath)
	if err != nil {
		return fmt.Errorf("cannot resolve symlink for executable: %w", err)
	}

	if err := replaceBinary(binaryPath, selfPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Printf("  ✓ Updated to %s\n", target)
	return nil
}

func latestReleaseTag() (string, error) {
	rel, err := fetchReleaseFunc("latest")
	if err != nil {
		return "", err
	}
	return rel.TagName, nil
}

func fetchRelease(tagOrLatest string) (*ghRelease, error) {
	var url string
	if tagOrLatest == "latest" {
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", githubRepo)
	} else {
		url = fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", githubRepo, tagOrLatest)
	}

	resp, err := http.Get(url) //nolint:noctx // simple CLI tool, no context needed
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck // read-only response body
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}
	var rel ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func assetURL(rel *ghRelease, name string) string {
	for _, a := range rel.Assets {
		if a.Name == name {
			return a.BrowserDownloadURL
		}
	}
	return ""
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url) //nolint:noctx // simple CLI tool
	if err != nil {
		return err
	}
	defer resp.Body.Close() //nolint:errcheck // response body
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d fetching %s", resp.StatusCode, url)
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck // write file
	_, err = io.Copy(f, resp.Body)
	return err
}

func verifyChecksum(archivePath, checksumPath string) error {
	data, err := os.ReadFile(checksumPath)
	if err != nil {
		return fmt.Errorf("cannot read checksums.txt: %w", err)
	}
	filename := filepath.Base(archivePath)
	expected := ""
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == filename {
			expected = fields[0]
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("no checksum entry found for %s", filename)
	}

	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck // read-only
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := fmt.Sprintf("%x", h.Sum(nil))
	if actual != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}

func extractBinary(archivePath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck // read-only

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close() //nolint:errcheck // gzip reader

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Name == "skill-auditor" || filepath.Base(hdr.Name) == "skill-auditor" {
			out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil { //nolint:gosec // archive is checksum-verified
				out.Close() //nolint:errcheck // best-effort
				return err
			}
			return out.Close()
		}
	}
	return fmt.Errorf("skill-auditor binary not found inside archive")
}

func replaceBinary(src, dest string) error {
	dir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(dir, ".skill-auditor-update-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()

	srcF, err := os.Open(src)
	if err != nil {
		tmp.Close()        //nolint:errcheck // best-effort
		os.Remove(tmpPath) //nolint:errcheck // best-effort
		return err
	}
	defer srcF.Close() //nolint:errcheck // read-only

	if _, err := io.Copy(tmp, srcF); err != nil {
		tmp.Close()        //nolint:errcheck // best-effort
		os.Remove(tmpPath) //nolint:errcheck // best-effort
		return err
	}
	if err := tmp.Chmod(0o755); err != nil {
		tmp.Close()        //nolint:errcheck // best-effort
		os.Remove(tmpPath) //nolint:errcheck // best-effort
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath) //nolint:errcheck // best-effort
		return err
	}
	return os.Rename(tmpPath, dest)
}
