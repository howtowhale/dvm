package config

type DvmOptions struct {
	DvmDir             string
	MirrorURL          string
	Token              string
	Shell              string
	Debug              bool
	Silent             bool
	IncludePrereleases bool
}
