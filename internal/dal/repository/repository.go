package repository

// New creates repository by connection
func New(conn Connection, options *Options) (*Repository, error) {
	r, err := create(conn, options)
	if err != nil {
		return nil, err
	}
	return &Repository{
		adapter: r,
	}, nil
}
