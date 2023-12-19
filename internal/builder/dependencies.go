package builder

//go:generate mockery --name=File --outpkg=mockbuilder --output=./mockbuilder
type File interface {
	Exists(fileName string) (bool, error)
}
