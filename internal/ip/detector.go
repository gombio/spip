package ip

type Detector interface {
	GetIP() (string, error)
}