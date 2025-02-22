package remote

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"scribe/internal/compressed"
	"scribe/internal/config"
	"scribe/internal/diff"
	"scribe/internal/history"
	"scribe/internal/ignore"
	"scribe/internal/util"
	"strconv"
	"strings"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v3"
)

type Remote struct {
	SshClient  *ssh.Client
	SftpClient *sftp.Client
	Config     *config.Config
	WD         string
}

const (
	DirObjects = "objects"
	DirCommits = "commits"
	FileHead   = "HEAD"
)

func (r *Remote) LocalWD() string {
	return filepath.Dir(r.Config.Location)
}

func (r *Remote) Close() error {
	if r == nil {
		return nil
	}

	if r.SftpClient != nil {
		if err := r.SftpClient.Close(); err != nil {
			_ = r.SshClient.Close()
			return errors.Join(errors.New("failed to close SFTP client"), err)
		}
		r.SftpClient = nil
	}

	if r.SshClient != nil {
		if err := r.SshClient.Close(); err != nil {
			return errors.Join(errors.New("failed to close SSH client"), err)
		}
		r.SshClient = nil
	}

	return nil
}

func Connect(c *config.Config) (*Remote, error) {
	if c == nil {
		return nil, errors.New("cannot connect, config is nil")
	}

	r := &Remote{Config: c, WD: c.Path}
	var err error

	r.SshClient, err = connectSsh(c)
	if err != nil {
		return nil, errors.Join(errors.New("failed to establish ssh connection"), err)
	}

	r.SftpClient, err = sftp.NewClient(r.SshClient)
	if err != nil {
		return nil, errors.Join(errors.New("failed to establish sftp connection"), err)
	}

	if err := r.SftpClient.MkdirAll(c.Path); err != nil && !os.IsExist(err) {
		return nil, errors.Join(errors.New("failed to ensure path exists"), err)
	}

	return r, nil
}

func (r *Remote) Mkdir(p string) error {
	if err := r.SftpClient.MkdirAll(path.Join(r.WD, p)); err != nil && !os.IsExist(err) {
		return errors.Join(errors.New("failed to create directory"), err)
	}
	return nil
}

func (r *Remote) Write(f *os.File, p string) error {
	if err := r.SftpClient.MkdirAll(path.Join(r.WD, filepath.Dir(p))); err != nil && !os.IsExist(err) {
		return errors.Join(errors.New("failed to create parent directories"), err)
	}
	rf, err := r.SftpClient.Create(path.Join(r.WD, p))
	if err != nil {
		return errors.Join(errors.New("failed to create remote file"), err)
	}
	_, err = compressed.Write(f, rf)
	if err != nil {
		return errors.Join(errors.New("failed to write compressed data"), err)
	}
	return nil
}

func (r *Remote) Read(remote string, local string) error {
	if err := os.MkdirAll(path.Join(r.LocalWD(), filepath.Dir(local)), 0764); err != nil && !os.IsExist(err) {
		return errors.Join(errors.New("failed to create parent directories"), err)
	}

	f, err := os.Create(path.Join(r.LocalWD(), local))
	if err != nil {
		return errors.Join(errors.New("failed to create file"), err)
	}
	defer f.Close()

	rf, err := r.SftpClient.Open(path.Join(r.WD, remote))
	if err != nil {
		return errors.Join(errors.New("failed to open remote file"), err)
	}
	defer rf.Close()

	_, err = compressed.Read(rf, f)
	if err != nil {
		return errors.Join(errors.New("failed to write compressed data"), err)
	}

	return nil
}

func (r *Remote) CommitFile(f *os.File, path string, c *history.Commit) error {
	cf := history.CommitFile{Path: path}

	if h, err := util.HashReader(f); err != nil {
		return errors.Join(errors.New("failed to calculate file hash"), err)
	} else {
		cf.Hash = h
	}

	if has, err := r.HasObject(cf.Hash); err != nil {
		return errors.Join(errors.New("failed to check object existence"), err)
	} else if !has {
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return errors.Join(errors.New("failed to seek file to start"), err)
		}
		if err := r.WriteObject(f, cf.Hash); err != nil {
			return errors.Join(errors.New("failed to write object"), err)
		}
	}

	c.Files = append(c.Files, cf)
	return nil
}

func hashToObjectPath(h string) string {
	return h[:1] + "/" + h[1:2] + "/" + h[2:8] + "/" + h[8:]
}

func (r *Remote) HasObject(h string) (bool, error) {
	_, err := r.SftpClient.Stat(path.Join(r.WD, DirObjects, hashToObjectPath(h)))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.Join(errors.New("failed to stat object"), err)
	}

	return true, nil
}

func (r *Remote) WriteObject(f *os.File, h string) error {
	if err := r.Write(f, path.Join(DirObjects, hashToObjectPath(h))); err != nil {
		return errors.Join(errors.New("failed to write object file"), err)
	}
	return nil
}

func (r *Remote) ReadObject(cf history.CommitFile) error {
	if err := r.Read(path.Join(DirObjects, hashToObjectPath(cf.Hash)), cf.Path); err != nil {
		return errors.Join(errors.New("failed to read object file"), err)
	}
	return nil
}

func (r *Remote) WriteCommit(f *os.File, c *history.Commit) error {
	if err := r.Write(f, path.Join(DirCommits, c.FileName())); err != nil {
		return errors.Join(errors.New("failed to write commit file"), err)
	}
	return nil
}

func (r *Remote) SetHeadCommit(c *history.Commit) error {
	rf, err := r.SftpClient.Create(path.Join(r.WD, FileHead))
	if err != nil {
		return errors.Join(errors.New("failed to create head file on remote"), err)
	}
	defer rf.Close()

	_, err = fmt.Fprintf(rf, "%d", c.Created)
	if err != nil {
		return errors.Join(errors.New("failed to write commit to head file on remote"), err)
	}
	return nil
}

func (r *Remote) RepoIsEmpty() (bool, error) {
	fi, err := r.SftpClient.Stat(r.WD)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil
		}
		return false, errors.Join(errors.New("failed to read repo remote dir info "+r.WD), err)
	}

	if !fi.IsDir() {
		return false, fmt.Errorf("remote dir path exists but is not a directory: %s", r.WD)
	}

	fis, err := r.SftpClient.ReadDir(r.WD)
	if err != nil {
		return false, errors.Join(errors.New("failed to read repo remote dir contents "+r.WD), err)
	}

	return len(fis) == 0, nil
}

func (r *Remote) Commit(msg string) error {
	commit := &history.Commit{
		Message: msg,
		Ignore:  r.Config.Ignore,
	}

	m := ignore.GetMatcher(r.Config)

	localWd := r.LocalWD()

	if err := filepath.WalkDir(localWd, func(absPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		repoPath, err := filepath.Rel(localWd, absPath)
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
		f, err := os.Open(absPath)
		if err != nil {
			return errors.Join(errors.New("failed to open file"), err)
		}
		defer f.Close()
		return r.CommitFile(f, strings.Join(gitPath, "/"), commit)
	}); err != nil {
		return errors.Join(fmt.Errorf("failed to walk repo dir %s", localWd), err)
	}

	if err := commit.Save(); err != nil {
		return errors.Join(errors.New("failed to save commit"), err)
	}

	if cf, err := commit.Open(); err != nil {
		return errors.Join(errors.New("failed to open commit file"), err)
	} else {
		defer cf.Close()
		if err := r.WriteCommit(cf, commit); err != nil {
			return errors.Join(errors.New("failed to write commit"), err)
		}
	}

	if err := r.SetHeadCommit(commit); err != nil {
		return errors.Join(errors.New("failed to set commit as head"), err)
	}

	return nil
}

func (r *Remote) InitialCommit() error {
	wd, err := os.Getwd()
	if err != nil {
		return errors.Join(errors.New("failed to get working directory"), err)
	}

	m := ignore.GetMatcher(r.Config)
	commit := &history.Commit{
		Message: "init",
		Ignore:  r.Config.Ignore,
	}

	if err := filepath.WalkDir(wd, func(absPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		repoPath, err := filepath.Rel(r.LocalWD(), absPath)
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
		f, err := os.Open(absPath)
		if err != nil {
			return errors.Join(errors.New("failed to open file"), err)
		}
		defer f.Close()
		return r.CommitFile(f, strings.Join(gitPath, "/"), commit)
	}); err != nil {
		return errors.Join(errors.New("failed to walk repo dir"), err)
	}

	if err := commit.Save(); err != nil {
		return errors.Join(errors.New("failed to save commit"), err)
	}

	if cf, err := commit.Open(); err != nil {
		return errors.Join(errors.New("failed to open commit file"), err)
	} else {
		defer cf.Close()
		if err := r.WriteCommit(cf, commit); err != nil {
			return errors.Join(errors.New("failed to write commit"), err)
		}
	}

	if err := r.SetHeadCommit(commit); err != nil {
		return errors.Join(errors.New("failed to set initial commit as head"), err)
	}

	r.Config.Commit = commit.Created
	if err := r.Config.Save(); err != nil {
		return errors.Join(errors.New("failed to save config"), err)
	}

	return nil
}

func (r *Remote) PullCommits() error {
	fileInfos, err := r.SftpClient.ReadDir(path.Join(r.WD, DirCommits))
	if err != nil {
		return errors.Join(errors.New("failed to read commits directory on remote"), err)
	}

	for _, fileInfo := range fileInfos {
		name := fileInfo.Name()
		if fileInfo.IsDir() || !strings.HasSuffix(name, ".yaml") {
			continue
		}

		if err := r.Read(path.Join(DirCommits, name), path.Join(".scribe", name)); err != nil {
			return errors.Join(fmt.Errorf("failed to read remote file %s", name), err)
		}
	}

	return nil
}

func (r *Remote) GetHeadCommit() (*history.Commit, error) {
	rf, err := r.SftpClient.Open(path.Join(r.WD, FileHead))
	if err != nil {
		return nil, errors.Join(errors.New("failed to open head file"), err)
	}
	defer rf.Close()

	cb, err := io.ReadAll(rf)
	if err != nil {
		return nil, errors.Join(errors.New("failed to read head file from remote"), err)
	}
	ci, err := strconv.ParseInt(string(cb), 10, 64)
	if err != nil {
		return nil, errors.Join(errors.New("failed to read head commit number from remote"), err)
	}
	cf, err := os.Open(filepath.Join(r.LocalWD(), ".scribe", fmt.Sprintf("%x.yaml", ci)))
	if err != nil {
		return nil, errors.Join(errors.New("failed to open head commit file locally"), err)
	}
	defer cf.Close()

	c := &history.Commit{}
	yd := yaml.NewDecoder(cf)
	if err := yd.Decode(c); err != nil {
		return nil, errors.Join(errors.New("failed to decode local commit file"), err)
	}

	return c, nil
}

func (r *Remote) CloneCommit(c *history.Commit) error {
	r.Config.Ignore = c.Ignore

	// get files from remote
	for _, f := range c.Files {
		if err := r.ReadObject(f); err != nil {
			return errors.Join(errors.New("failed to read object from remote"), err)
		}
	}

	r.Config.Commit = c.Created
	return r.Config.Save()
}

func (r *Remote) CheckoutCommit(c *history.Commit) error {
	if r.Config.Commit == c.Created {
		return nil
	}

	r.Config.Ignore = c.Ignore

	currentCommit, err := r.Config.CurrentCommit()
	if err != nil {
		return errors.Join(errors.New("failed to get current commit"), err)
	}

	var locallyChanged diff.DiffList
	if currentCommit.Created != c.Created {
		locallyChanged, err = diff.LocalFromCommit(r.Config, currentCommit)
		if err != nil {
			return errors.Join(errors.New("failed to diff local changes with current commit"), err)
		}
	}

	localWd := r.LocalWD()

	// get files from remote
	for _, f := range c.Files {
		// check if file exists
		ccf, exists := currentCommit.File(f.Path)
		if exists {
			// check if file has changed on remote
			if ccf.Hash == f.Hash {
				continue
			}
			// check if file has changed locally
			if locallyChanged.HasModifyOrDelete(f.Path) {
				panic(fmt.Sprintf("checkout failed: conflict %s (local and remote modified/deleted)\n", f.Path))
			}
		} else {
			// locally also created?
			if locallyChanged.HasCreate(f.Path) {
				panic(fmt.Sprintf("checkout failed: conflict %s (local and remote created)\n", f.Path))
			}
		}
	}

	// delete files that shouldn't exist
	m := ignore.GetMatcher(r.Config)
	if err := filepath.WalkDir(localWd, func(absPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		repoPath, err := filepath.Rel(localWd, absPath)
		if err != nil {
			return err
		}
		repoPath = path.Clean(repoPath)

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

		repoUnixPath := strings.Join(gitPath, "/")

		if locallyChanged.HasCreate(repoUnixPath) {
			return nil
		}

		for _, f := range c.Files {
			if f.Path == repoPath {
				return nil
			}
		}

		log.Printf("delete %s\n", repoPath)
		return os.Remove(absPath)
	}); err != nil {
		return errors.Join(errors.New("error while walking local repo path"), err)
	}

	r.Config.Commit = c.Created
	return r.Config.Save()
}
