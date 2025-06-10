package builder

//go:generate mockery
type File interface {
	Exists(fileName string) (bool, error)
}
