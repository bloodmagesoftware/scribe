package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

const keyringService = "de.bloodmagesoftware.scribe"

var DefaultIgnore = []string{
	".DS_Store",
	".vs",
	".idea",
	".vscode",
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
	"*.xcodeproj",
	"*.xcworkspace",
	"*.sln",
	"*.suo",
	"*.opensdf",
	"*.sdf",
	"*.VC.db",
	"*.VC.opendb",
	"SourceArt/**/*.png",
	"SourceArt/**/*.tga",
	"Binaries/*",
	"Plugins/**/Binaries/*",
	"Build/*",
	"!Build/*/",
	"Build/*/**",
	"!Build/*/PakBlacklist*.txt",
	"!Build/**/*.ico",
	"*_BuiltData.uasset",
	"Saved/*",
	"Intermediate/*",
	"Plugins/**/Intermediate/*",
	"DerivedDataCache/*",
}

type Config struct {
	Host     string   `yaml:"host"`
	Port     int      `yaml:"port"`
	User     string   `yaml:"user"`
	Password string   `yaml:"-"`
	Path     string   `yaml:"path"`
	Ignore   []string `yaml:"ignore"`
}

func (c *Config) fullUser() string {
	return fmt.Sprintf("%s@%s:%d", c.User, c.Host, c.Port)
}

func (c *Config) Save() error {
	var f *os.File
	{
		var err error
		f, err = os.OpenFile(".scribe.yaml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return errors.Join(errors.New("failed to open file .scribe.yaml"), err)
		}
	}

	ye := yaml.NewEncoder(f)
	ye.SetIndent(4)
	if err := ye.Encode(c); err != nil {
		return errors.Join(errors.New("failed to yaml encode into file .scribe.yaml"), err)
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
	var f *os.File
	{
		var err error
		f, err = os.OpenFile(".scribe.yaml", os.O_RDONLY, 0600)
		if err != nil {
			return nil, errors.Join(errors.New("failed to open file .scribe.yaml"), err)
		}
	}

	yd := yaml.NewDecoder(f)
	c := &Config{}
	if err := yd.Decode(c); err != nil {
		return nil, errors.Join(errors.New("failed to yaml decode from file .scribe.yaml"), err)
	}

	{
		var err error
		if c.Password, err = keyring.Get(keyringService, c.fullUser()); err != nil {
			return nil, errors.Join(errors.New("failed to get keyring credentials"), err)
		}
	}

	return c, nil
}
