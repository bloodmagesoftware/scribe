package diff

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"scribe/internal/config"
	"scribe/internal/history"
	"scribe/internal/ignore"
	"scribe/internal/util"
	"strings"
)

type (
	DiffList []Diff
	Diff     struct {
		Path string
		Type DiffType
	}

	DiffType uint8
)

const (
	DiffTypeCreate DiffType = iota
	DiffTypeDelete
	DiffTypeModify
)

func LocalFromCommit(conf *config.Config, c *history.Commit) (DiffList, error) {
	root := filepath.Dir(conf.Location)
	var diff DiffList

	for _, cf := range c.Files {
		lfp := filepath.Join(root, cf.Path)

		// does file exist?
		fi, err := os.Stat(lfp)
		if err != nil {
			if os.IsNotExist(err) {
				diff = append(diff, Diff{cf.Path, DiffTypeDelete})
				continue
			}
			return nil, errors.Join(errors.New("failed to read file"), err)
		}
		if fi.IsDir() {
			diff = append(diff, Diff{cf.Path, DiffTypeDelete})
			continue
		}

		// has file changed?
		f, err := os.Open(lfp)
		if err != nil {
			return nil, errors.Join(errors.New("failed to open file"), err)
		}
		defer f.Close()
		h, err := util.HashReader(f)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("failed to hash file %s", lfp), err)
		}
		if h != cf.Hash {
			diff = append(diff, Diff{cf.Path, DiffTypeModify})
		}
	}

	m := ignore.GetMatcher(conf)
	if err := filepath.WalkDir(root, func(absPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		repoPath, err := filepath.Rel(root, absPath)
		if err != nil {
			return errors.Join(errors.New("failed to get relative path"), err)
		}
		gitPath := util.TrimSliceEmptyString(strings.Split(repoPath, string(filepath.Separator)))
		isDir := d.IsDir()
		if m.Match(gitPath, isDir) {
			// excluded from ignore
			if isDir {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		if isDir {
			return nil
		}
		repoPosixPath := path.Clean(strings.Join(gitPath, "/"))
		for _, cf := range c.Files {
			if cf.Path == repoPosixPath {
				return nil
			}
		}

		diff = append(diff, Diff{repoPosixPath, DiffTypeCreate})
		return nil
	}); err != nil {
		return nil, errors.Join(fmt.Errorf("error while walking dir %s", root), err)
	}

	return diff, nil
}

func (dl DiffList) HasDelete(path string) bool {
	for _, d := range dl {
		if d.Type == DiffTypeDelete {
			return d.Path == path
		}
	}
	return false
}

func (dl DiffList) HasModify(path string) bool {
	for _, d := range dl {
		if d.Type == DiffTypeModify {
			return d.Path == path
		}
	}
	return false
}

func (dl DiffList) HasModifyOrDelete(path string) bool {
	for _, d := range dl {
		if d.Type == DiffTypeModify || d.Type == DiffTypeDelete {
			return d.Path == path
		}
	}
	return false
}

func (dl DiffList) HasCreate(path string) bool {
	for _, d := range dl {
		if d.Type == DiffTypeCreate {
			return d.Path == path
		}
	}
	return false
}
