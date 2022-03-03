package blorg

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBlorg(t *testing.T) {
	// Re-generate this file with `find testdata/public -type f | sort -u | xargs md5sum > testdata/public.md5`
	hashFile, err := os.Open("testdata/public.md5")
	if err != nil {
		t.Errorf("Could not open hash file: %s", err)
		return
	}
	defer hashFile.Close()
	scanner := bufio.NewScanner(hashFile)
	committedHashes := make(map[string]string)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) != 2 {
			t.Errorf("Could not split hash entry line in 2: len(parts)=%d", len(parts))
			return
		}
		hash := parts[0]
		fileName := parts[1]
		committedHashes[fileName] = hash
	}
	if err := scanner.Err(); err != nil {
		t.Errorf("Failed to read hash file: %s", err)
		return
	}

	config, err := ReadConfig("testdata/blorg.org")
	if err != nil {
		t.Errorf("Could not read config: %s", err)
		return
	}
	if err := config.Render(); err != nil {
		t.Errorf("Could not render: %s", err)
		return
	}

	renderedFileHashes := make(map[string]string)
	err = filepath.WalkDir(config.PublicDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		hash := md5.Sum(data)
		renderedFileHashes[path] = hex.EncodeToString(hash[:])
		return nil
	})
	if err != nil {
		t.Errorf("Could not determine hashes of rendered files: %s", err)
		return
	}

	for file, rendered := range renderedFileHashes {
		if _, ok := committedHashes[file]; !ok {
			t.Errorf("New file %s does not have a committed hash", file)
			continue
		}
		committed := committedHashes[file]
		committedHashes[file] = "" // To check if there are missing files later.
		if rendered != committed {
			t.Errorf("PublicDir hashes do not match for %s: '%s' -> '%s'", file, committed, rendered)
		}
	}
	for file, committed := range committedHashes {
		if committed != "" {
			t.Errorf("Missing file %s has a committed hash, but was not rendered", file)
		}
	}
}
