package cmd

import (
	"github.com/cnupp/cli/pkg"
)

func GitRemote(appID, remote string) error {
	c, appID, err := load(appID)

	if err != nil {
		return err
	}

	return git.CreateRemote(c.GitHost(), remote, appID)
}
