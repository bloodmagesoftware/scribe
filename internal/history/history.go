package history

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"scribe/internal/util"
	"time"

	"gopkg.in/yaml.v3"
)

type (
	History []Commit

	Commit struct {
		Created int64        `yaml:"created_at"`
		Files   []CommitFile `yaml:"files"`
		Message string       `yaml:"message"`
		Ignore  string       `yaml:"ignore"`
		fp      string       `yaml:"-"`
	}

	CommitFile struct {
		Path string
		Hash string
	}
)

const (
	HistoryDirName = ".scribe"
)

func findHistoryDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		fp := filepath.Join(wd, HistoryDirName)
		if util.Exists(fp) {
			return fp, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd || parent == "/" || parent == "." {
			break
		}
	}
	return "", errors.New("no " + HistoryDirName + "/ found")
}

func Init() error {
	if err := os.Mkdir(HistoryDirName, 0764); err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func (c *Commit) Save() error {
	if c.Created == 0 {
		c.Created = time.Now().Unix()
	}

	if len(c.fp) == 0 {
		hdp, err := findHistoryDir()
		if err != nil {
			return err
		}
		c.fp = filepath.Join(hdp, fmt.Sprintf("%x.yaml", c.Created))
	}

	f, err := os.Create(c.fp)
	if err != nil {
		return err
	}

	ye := yaml.NewEncoder(f)
	defer ye.Close()
	return ye.Encode(c)
}

func (c *Commit) Open() (*os.File, error) {
	if c.Created == 0 {
		c.Created = time.Now().Unix()
	}

	if len(c.fp) == 0 {
		hdp, err := findHistoryDir()
		if err != nil {
			return nil, err
		}
		c.fp = filepath.Join(hdp, fmt.Sprintf("%x.yaml", c.Created))
	}

	return os.Open(c.fp)
}

func (c *Commit) FileName() string {
	if c.Created == 0 {
		c.Created = time.Now().Unix()
	}
	return fmt.Sprintf("%x.yaml", c.Created)
}

func (c *Commit) File(name string) (CommitFile, bool) {
	for _, f := range c.Files {
		if f.Path == name {
			return f, true
		}
	}
	return CommitFile{}, false
}
