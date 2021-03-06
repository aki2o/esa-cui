package action

import (
	"errors"
	"path/filepath"
	"fmt"
	"os"
	log "github.com/sirupsen/logrus"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type EsaCuiActionContext struct {
	post_strage_root_path	string
	post_body_strage_path	string
	User                    esa.User
	Team					string
	Cwd						string
	Client					*esa.Client
	PecoPreferred           bool
}

func (c *EsaCuiActionContext) Root() string {
	return filepath.Join(c.post_strage_root_path, c.Team)
}

func (c *EsaCuiActionContext) BodyRoot() string {
	return filepath.Join(c.post_body_strage_path, c.Team)
}

var Context *EsaCuiActionContext

func SetupContext(team string, access_token string, login bool) error {
	Context = &EsaCuiActionContext{}

	if team == "" {
		return errors.New("Invalid Team!")
	}
	
	Context.post_strage_root_path	= filepath.Join(util.LocalRootPath(), ".posts")
	Context.post_body_strage_path	= filepath.Join(util.LocalRootPath(), "posts")
	Context.Team					= team
	Context.Cwd						= Context.Root()
	Context.Client					= esa.NewClient(access_token)

	if login {
		var err error
		
		Context.User, err = Context.Client.User.Get()
		if err != nil {
			log.WithFields(log.Fields{ "team": Context.Team }).Error("Failed to fetch login user")
			fmt.Fprintln(os.Stderr, "Failed to fetch login user information!")
		}
	}
	
	log.WithFields(log.Fields{ "team": Context.Team, "cwd": Context.Cwd }).Debug("setup Context")
	
	return nil
}
