package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"scribe/internal/util"

	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

const (
	keyringService = "de.bloodmagesoftware.scribe"
	ConfigFileName = ".scribe.yaml"
)

var DefaultIgnore = []string{
	".DS_Store",
	".vs/",
	".idea/",
	".vscode/",
	".git/",
	"*.slo",
	"*.lo",
	"*.o",
	"*.obj",
	"*.gch",
	"*.pch",
	"*.so",
	"*.dylib",
	"*.dll",
	"*.mod",
	"*.lai",
	"*.la",
	"*.a",
	"*.lib",
	"*.exe",
	"*.out",
	"*.app",
	"*.ipa",
}

type Config struct {
	Version  uint8    `yaml:"version"`
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	User     string   `yaml:"user"`
	Password string   `yaml:"-"`
	Path     string   `yaml:"path"`
	Commit   int64    `yaml:"commit"`
	Ignore   []string `yaml:"ignore"`
	Location string   `yaml:"-"`
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
	}
	return "", errors.New("no " + ConfigFileName + " found")
}

func (c *Config) fullUser() string {
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
	c.Version = 1

	ye := yaml.NewEncoder(f)
	ye.SetIndent(4)
	if err := ye.Encode(c); err != nil {
		return errors.Join(errors.New("failed to yaml encode into file "+ConfigFileName), err)
	}

	if err := keyring.Set(
		keyringService,
		c.fullUser(),
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
		keyringService,
		c.fullUser(),
		c.Password,
	); err != nil {
		return errors.Join(errors.New("failed to set keyring credentials"), err)
	}

	return nil
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
		if c.Password, err = keyring.Get(keyringService, c.fullUser()); err != nil {
			return nil, errors.Join(errors.New("failed to get keyring credentials"), err)
		}
	}

	return c, nil
}
