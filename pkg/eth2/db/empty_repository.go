package db

// notest
type EmptyRepository struct{}

func (er EmptyRepository) FirstOrCreate(Validator) (v Validator, e error) {
	return
}

func (er EmptyRepository) Update(Validator) error {
	return nil
}

func (er EmptyRepository) Validator(index uint) (v Validator, e error) {
	return
}

func (er EmptyRepository) Migrate() error {
	return nil
}
