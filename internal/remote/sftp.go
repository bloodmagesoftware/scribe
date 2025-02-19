package remote

import (
	"errors"
	"path"
	"scribe/internal/config"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Remote struct {
	conn   *ssh.Client
	client *sftp.Client
	wd     string
}

func (r *Remote) Close() error {
	if r == nil {
		return nil
	}

	if r.client != nil {
		if err := r.client.Close(); err != nil {
			_ = r.conn.Close()
			return err
		}
		r.client = nil
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			return err
		}
		r.conn = nil
	}

	return nil
}

func Connect(c *config.Config) (*Remote, error) {
	if c == nil {
		return nil, errors.New("cannot connect, config is nil")
	}

	r := &Remote{wd: c.Path}
	var err error

	r.conn, err = connectSsh(c)
	if err != nil {
		return nil, errors.Join(errors.New("failed to establish ssh connection"), err)
	}

	r.client, err = sftp.NewClient(r.conn)
	if err != nil {
		return nil, errors.Join(errors.New("failed to establish sftp connection"), err)
	}

	if err := r.client.MkdirAll(c.Path); err != nil {
		return nil, errors.Join(errors.New("failed to ensure path exists"), err)
	}

	return r, nil
}

func (r *Remote) Mkdir(p string) error {
	return r.client.MkdirAll(path.Join(r.wd, p))
}
