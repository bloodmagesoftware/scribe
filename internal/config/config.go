package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"scribe/internal/history"
	"scribe/internal/util"

	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

const (
	KeyringService = "de.bloodmagesoftware.scribe"
	ConfigFileName = ".scribe.yaml"
)

const Version = 1

const DefaultIgnore = `.DS_Store
.vs/
.idea/
.vscode/
.git/
.gitattributes
.gitignore
`

type Config struct {
	Version  uint8  `yaml:"version"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"-"`
	Path     string `yaml:"path"`
	Commit   int64  `yaml:"commit"`
	Ignore   string `yaml:"ignore"`
	Location string `yaml:"-"`
}

func findConfigFile() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		fp := filepath.Join(wd, ConfigFileName)
		if util.Exists(fp) {
			return fp, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd || parent == "/" || parent == "." {
			break
		}
		wd = parent
	}
	return "", errors.New("no " + ConfigFileName + " found")
}

func (c *Config) FullUser() string {
	return fmt.Sprintf("%s@%s:%d", c.User, c.Host, c.Port)
}

func (c *Config) SaveNew() error {
	var f *os.File
	{
		var err error
		f, err = os.Create(ConfigFileName)
		if err != nil {
			return errors.Join(errors.New("failed to open file "+ConfigFileName), err)
		}
	}
	c.Location, _ = filepath.Abs(f.Name())
	c.Version = Version

	ye := yaml.NewEncoder(f)
	ye.SetIndent(4)
	if err := ye.Encode(c); err != nil {
		return errors.Join(errors.New("failed to yaml encode into file "+ConfigFileName), err)
	}

	if err := keyring.Set(
		KeyringService,
		c.FullUser(),
		c.Password,
	); err != nil {
		return errors.Join(errors.New("failed to set keyring credentials"), err)
	}

	return nil
}

func (c *Config) Save() error {
	f, err := os.Create(c.Location)
	if err != nil {
		return errors.Join(errors.New("failed to open file "+ConfigFileName), err)
	}

	ye := yaml.NewEncoder(f)
	ye.SetIndent(4)
	if err := ye.Encode(c); err != nil {
		return errors.Join(errors.New("failed to yaml encode into file "+ConfigFileName), err)
	}

	if err := keyring.Set(
		KeyringService,
		c.FullUser(),
		c.Password,
	); err != nil {
		return errors.Join(errors.New("failed to set keyring credentials"), err)
	}

	return nil
}

func RepoRoot() (string, error) {
	cfp, err := findConfigFile()
	if err != nil {
		return "", errors.Join(errors.New("failed to find config file"), err)
	}
	return filepath.Dir(cfp), nil
}

func Load() (*Config, error) {
	cfp, err := findConfigFile()
	if err != nil {
		return nil, errors.Join(errors.New("failed to find config file"), err)
	}

	var f *os.File
	f, err = os.Open(cfp)
	if err != nil {
		return nil, errors.Join(errors.New("failed to open file "+ConfigFileName), err)
	}
	defer f.Close()

	yd := yaml.NewDecoder(f)
	c := &Config{Location: cfp}
	if err := yd.Decode(c); err != nil {
		return nil, errors.Join(errors.New("failed to yaml decode from file "+ConfigFileName), err)
	}

	{
		var err error
		if c.Password, err = keyring.Get(KeyringService, c.FullUser()); err != nil {
			return nil, errors.Join(errors.New("failed to get keyring credentials"), err)
		}
	}

	return c, nil
}

func (c *Config) CurrentCommit() (*history.Commit, error) {
	if c.Commit == 0 {
		return nil, errors.New("no commit checked out")
	}
	f, err := os.Open(filepath.Join(filepath.Dir(c.Location), ".scribe", fmt.Sprintf("%x.yaml", c.Commit)))
	if err != nil {
		return nil, errors.Join(fmt.Errorf("failed to open commit file for commit %x", c.Commit), err)
	}
	defer f.Close()

	yd := yaml.NewDecoder(f)
	commit := &history.Commit{}
	if err := yd.Decode(commit); err != nil {
		return nil, errors.Join(errors.New("failed to decode commit file"), err)
	}

	return commit, nil
}
