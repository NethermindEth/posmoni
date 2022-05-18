package db

type Repository interface {
	FirstOrCreate(Validator) (Validator, error)
	Update(Validator) error
	Validator(index uint) (Validator, error)
	Migrate() error
}
