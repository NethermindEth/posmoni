package db

import "gorm.io/gorm"

// notest

type ValidatorORM struct {
	gorm.Model
	Validator
}

type SQLiteRepository struct {
	DB *gorm.DB
}

func (r *SQLiteRepository) FirstOrCreate(v Validator) (Validator, error) {
	var m ValidatorORM
	if err := r.DB.First(&m, "idx = ?", v.Idx).Session(&gorm.Session{}).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return Validator{}, err
		}
		m = ValidatorORM{Validator: v}
		if err := r.DB.Create(&m).Session(&gorm.Session{}).Error; err != nil {
			return Validator{}, err
		}
	}
	return m.Validator, nil
}

func (r *SQLiteRepository) Update(v Validator) error {
	var m ValidatorORM
	if err := r.DB.Where("idx = ?", v.Idx).First(&m).Session(&gorm.Session{}).Error; err != nil {
		return err
	}

	m.Balance = v.Balance
	m.MissedAtts = v.MissedAtts
	m.MissedAttsTotal = v.MissedAttsTotal

	return r.DB.Save(&m).Error
}

func (r *SQLiteRepository) Validator(index uint) (Validator, error) {
	var m ValidatorORM
	if err := r.DB.First(&m, "idx = ?", index).Session(&gorm.Session{}).Error; err != nil {
		return Validator{}, err
	}
	return m.Validator, nil
}

func (r *SQLiteRepository) Migrate() error {
	return r.DB.AutoMigrate(&ValidatorORM{})
}
