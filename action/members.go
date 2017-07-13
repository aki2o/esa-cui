package action

import (
	"flag"
	"errors"
	"path/filepath"
	"encoding/json"
	"os"
	"bufio"
	"strings"
	"io/ioutil"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esa-cui/util"
)

type members struct {
	refresh bool
	name_print bool
	email_print bool
}

func init() {
	addProcessor(&members{}, "members", "Print members.")
}

func (self *members) SetOption(flagset *flag.FlagSet) {
	flagset.BoolVar(&self.refresh, "r", false, "Exec with ignore cache.")
	flagset.BoolVar(&self.name_print, "name", false, "Print name.")
	flagset.BoolVar(&self.email_print, "email", false, "Print email.")
}

func (self *members) Do(args []string) error {
	members, err := self.load()
	if err != nil { return err }

	if self.refresh || len(members) == 0 {
		if self.refresh {
			err = os.Remove(self.GetLocalStragePath())
			if err != nil { return err }
		}
		
		if err = self.fetch(); err != nil { return err }

		members, err = self.load()
		if err != nil { return err }
	}

	for _, member := range members {
		var name string = ""
		var email string = ""

		if self.name_print { name = member.Name }
		if self.email_print { email = member.Email }
		
		fmt.Printf("%-30s%-15s%s\n", member.ScreenName, name, email)
	}
	
	return nil
}

func (self *members) load() ([]*esa.Member, error) {
	var members = []*esa.Member{}
	
	if err := util.EnsureDir(self.GetLocalStragePath()); err != nil { return members, err }
	
	for _, node := range util.GetNodes(self.GetLocalStragePath()) {
		if node.IsDir() { continue }

		node_name_parts := strings.Split(node.Name(), ".")
		if len(node_name_parts) != 2 { continue }

		member_screen_name	:= node_name_parts[0]
		file_extension		:= node_name_parts[1]
		if file_extension != "json" { continue }

		bytes, err := ioutil.ReadFile(filepath.Join(self.GetLocalStragePath(), node.Name()))
		if err != nil { return members, err }
		
		var member esa.Member
		if err := json.Unmarshal(bytes, &member); err != nil {
			log.WithFields(log.Fields{ "name": node.Name() }).Error("Failed to load member")
			util.PutError(errors.New("Failed to load member data of "+member_screen_name+"!"))
		} else {
			members = append(members, &member)
		}
	}
	
	return members, nil
}

func (self *members) fetch() error {
	if err := util.EnsureDir(self.GetLocalStragePath()); err != nil { return err }

	log.Info("start to fetch member "+Context.Team)
	members, err := Context.Client.Members.Get(Context.Team)
	if err != nil { return err }

	for _, member := range members {
		self.StoreMemberData(&member)
	}

	log.Info("finished to fetch member "+Context.Team)
	return nil
}

func (self *members) StoreMemberData(member *esa.Member) {
	bytes, err := json.MarshalIndent(member, "", "\t")
	if err != nil {
		util.PutError(err)
		return
	}
	
	fp, err := os.Create(filepath.Join(self.GetLocalStragePath(), member.ScreenName+".json"))
	if err != nil { panic(err) }
	defer fp.Close()
	writer := bufio.NewWriter(fp)
	_, err = writer.Write(bytes)
	if err != nil { panic(err) }
	writer.Flush()
}

func (self *members) GetLocalStragePath() string {
	return filepath.Join(util.LocalRootPath(), "members", Context.Team)
}