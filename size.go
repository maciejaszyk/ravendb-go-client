package ravendb

// Size describes size of entity on disk
type Size struct {
	SizeInBytes int    `json:"SizeInBytes"`
	HumaneSize  string `json:"HumaneSize"`
}
