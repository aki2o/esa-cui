package util

import (
	"flag"
	"fmt"
	"os"
	"github.com/abiosoft/ishell"
)

type Processable interface {
	SetOption(flagset *flag.FlagSet)
	Do(args []string) error
}

type ProcessorRepository struct {
	processors map[string]Processable
	usages map[string]string
}

func (self *ProcessorRepository) ProcessorNames() []string {
	self.setup()
	
	var ret = make([]string, len(self.processors))
	
	index := 0
	for name, _ := range self.processors {
		ret[index] = name
		index += 1
	}
	
	return ret
}

func (self *ProcessorRepository) setup() {
	if self.processors == nil { self.processors = make(map[string]Processable) }
	if self.usages == nil { self.usages = make(map[string]string) }
}

func (self *ProcessorRepository) SetProcessor(name string, processor Processable) {
	self.setup()
	self.processors[name] = processor
}

func (self *ProcessorRepository) GetProcessor(name string) Processable {
	self.setup()
	return self.processors[name]
}

func (self *ProcessorRepository) SetUsage(name string, usage string) {
	self.setup()
	self.usages[name] = usage
}

func (self *ProcessorRepository) GetUsage(name string) string {
	self.setup()
	return self.usages[name]
}

type IshellAdapter struct {
	Processor Processable
	ProcessorName string
	ProcessorUsage string
}

func (self *IshellAdapter) Adapt(ctx *ishell.Context) {
	self.Run(ctx.Args)
}

func (self *IshellAdapter) Run(args []string) {
	flagset := flag.NewFlagSet(self.ProcessorName, flag.PanicOnError)
	
	var help_required bool = false
	flagset.BoolVar(&help_required, "h", false, "Show help.")
	
	self.Processor.SetOption(flagset)
	
	err := flagset.Parse(args)
	if err != nil {
		PutError(err)
		return
	}

	if help_required {
		fmt.Fprintf(os.Stderr, "%s: %s\n\nOptions:\n", self.ProcessorName, self.ProcessorUsage)
		flagset.PrintDefaults()
		return
	}
	
	err = self.Processor.Do(flagset.Args())
	if err != nil {
		PutError(err)
		return
	}
}
