package config

import (
	"io/ioutil"
	"log"
)

type DvmOptions struct {
	DvmDir             string
	MirrorURL          string
	Token              string
	Shell              string
	Debug              bool
	Silent             bool
	IncludePrereleases bool
	Logger             *log.Logger
}

func NewDvmOptions() DvmOptions {
	return DvmOptions{
		Logger: log.New(ioutil.Discard, "", log.LstdFlags),
	}
}
