package extractor

type Extractor interface {
	Extract(src, dest string) error
}
