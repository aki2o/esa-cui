package action

import (
	"fmt"
	"path/filepath"
	"strings"
	"encoding/json"
	"io"
	"bufio"
	"os"
	"time"
	"strconv"
	"errors"
	"regexp"
	log "github.com/sirupsen/logrus"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type ls struct {
	LongFormatRequired bool `short:"l" long:"long" description:"Print long format."`
	Recursive bool `short:"r" long:"recursive" description:"Exec recursively."`
	DirectoryOnly bool `short:"d" long:"directory" description:"Print only directory."`
	FileOnly bool `short:"f" long:"file" description:"Print only file."`
	writer io.Writer
}

func init() {
	registProcessor(func() util.Processable { return &ls{ writer: os.Stdout } }, "ls", "Print a list of category and post information.", "[OPTIONS]")
}

func (self *ls) Do(args []string) error {
	var path string = ""
	if len(args) > 0 { path = args[0] }
	
	self.printNodesIn(path, PhysicalPathOf(path))
	return nil
}

func (self *ls) printNodesIn(path string, physical_path string) {
	writer := bufio.NewWriter(self.writer)

	re, _ := regexp.Compile("/$")
	path = re.ReplaceAllString(path, "")+"/"
	
	for _, node := range util.GetNodes(physical_path) {
		node_physical_path := filepath.Join(physical_path, node.Name())
		
		if node.IsDir() {
			node_path := path+util.DecodePath(node.Name())
			
			if ! self.FileOnly {
				fmt.Fprintln(writer, self.makeDirLine(node_path))
				writer.Flush()
			}

			if self.Recursive { self.printNodesIn(node_path, node_physical_path) }
		} else if ! self.DirectoryOnly {
			var post esa.PostResponse
			
			post_number := node.Name()
			bytes, err := LoadPostData(post_number)
			
			if err == nil { err = json.Unmarshal(bytes, &post) }

			if err != nil {
				log.WithFields(log.Fields{ "name": node.Name(), "path": node_physical_path }).Error("Failed to load post")
				util.PutError(errors.New("Failed to load post data of "+post_number+"!"))
			} else {
				fmt.Fprintln(writer, self.makeFileLine(path, &post))
			}
		}
	}

	writer.Flush()
}

func (self *ls) makeDirLine(path string) string {
	return self.makePostStatPart(nil)+path+"/"
}

func (self *ls) makeFileLine(path string, post *esa.PostResponse) string {
	post_number := strconv.Itoa(post.Number)

	var name_part string
	if self.LongFormatRequired {
		var wip string = ""
		var lock string = ""
		var tag string = ""
		
		if post.Wip { wip = " [WIP]" }
		if _, err := os.Stat(GetPostLockPath(post_number)); err == nil { lock = " *Lock*" }
		if len(post.Tags) > 0 { tag = " #"+strings.Join(post.Tags, " #") }
		
		name_part = fmt.Sprintf("%s:%s%s %s%s", path+post_number, wip, lock, post.Name, tag)
	} else {
		name_part = fmt.Sprintf("%s: %s", path+post_number, post.Name)
	}

	return self.makePostStatPart(post)+name_part
}

func (self *ls) makePostStatPart(post *esa.PostResponse) string {
	if !self.LongFormatRequired { return "" }

	var create_user		string = ""
	var update_user		string = ""
	var post_size		string = ""
	var last_updated_at string = ""

	if post != nil {
		create_user = post.CreatedBy.ScreenName
		update_user = post.UpdatedBy.ScreenName

		file_info, err := os.Stat(GetPostBodyPath(strconv.Itoa(post.Number)))
		if err == nil {
			post_size = fmt.Sprintf("%d", file_info.Size())
		} else {
			post_size = "?"
		}
		
		updated_at, err := time.Parse("2006-01-02T15:04:05-07:00", post.UpdatedAt)
		if err != nil {
			last_updated_at = "** ** **:**"
		} else if updated_at.Year() == time.Now().Year() {
			last_updated_at = updated_at.Format("01 02 15:04")
		} else {
			last_updated_at = updated_at.Format("01 02  2006")
		}
	}
	
	return fmt.Sprintf(
		"%s %s %s %s ",
		fmt.Sprintf("%20s", create_user),
		fmt.Sprintf("%20s", update_user),
		fmt.Sprintf("%10s", post_size),
		fmt.Sprintf("%11s", last_updated_at),
	)
}
