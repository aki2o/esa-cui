package action

import (
	"regexp"
	"fmt"
	"github.com/aki2o/go-esa/esa"
	"github.com/aki2o/esal/util"
)

type uploadable struct {
	Wip bool `short:"w" long:"wip" description:"Update the post as wip."`
	Shipping bool `short:"s" long:"ship" description:"Ship the post."`
	Tags []string `short:"T" long:"tag" description:"Tag name labeling tha post." value-name:"TAG"`
	Category string `short:"C" long:"category" description:"Category of the post." value-name:"CATEGORY"`
	PostName string `short:"n" long:"name" description:"Name of the post." value-name:"NAME"`
	Message string `short:"m" long:"message" description:"Commit message." value-name:"MESSAGE"`
	TagsByPecoRequired bool `short:"t" long:"tagp" description:"Choice tags by peco."`
	CategoryByPecoRequired bool `short:"c" long:"categoryp" description:"Choice category by peco."`
	MessageByScan bool `short:"M" long:"message-by-scan" description:"Input commit message."`
}

func (self *uploadable) setWip(post *esa.Post, default_value bool) {
	wip := default_value
	
	if self.Wip { wip = true }
	if self.Shipping { wip = false }

	post.Wip = wip
}

func (self *uploadable) setTags(post *esa.Post, default_value []string) {
	tags := default_value
	
	if len(self.Tags) > 0 {
		tags = self.Tags
	} else if self.TagsByPecoRequired {
		if selected_tags, err := selectTagByPeco(); err == nil { tags = selected_tags }
	}

	post.Tags = tags
	self.Tags = tags
}

func (self *uploadable) setCategory(post *esa.Post, default_value string) {
	category := default_value
	re, _ := regexp.Compile("^/")
	
	if self.Category != "" {
		category = re.ReplaceAllString(self.Category, "")
	} else if self.CategoryByPecoRequired {
		categories, err := selectNodeByPeco("/"+ParentOf(CategoryOf(Context.Cwd)), true)
		if err == nil {
			category = categories[0]
			if category == "" { category = CategoryOf(Context.Cwd) }
			
			child_category := util.ScanString(fmt.Sprintf("Child Category of (%s): ", category))
			category = re.ReplaceAllString(category, "")+re.ReplaceAllString(child_category, "")
		}
	}

	post.Category = category
	self.Category = category
}

func (self *uploadable) setName(post *esa.Post, default_value string) {
	post_name := default_value
	
	if self.PostName != "" { post_name = self.PostName }

	post.Name = post_name
}

func (self *uploadable) setMessage(post *esa.Post) {
	message := "Update post."
	
	if self.Message != "" {
		message = self.Message
	} else if self.MessageByScan {
		message = util.ScanString("Commit Message: ")
	}

	post.Message = message
	self.Message = message
}
