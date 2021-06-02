package repo

import (
	"fmt"

	pq "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type RefreshRepoPostgres struct {
	db *gorm.DB
}

type temporalRecord struct {
	Name    []byte        `gorm:"primaryKey;index:,type:hash"`
	Heights pq.Int64Array `gorm:"type:integer[];index:,type:gin"`
}

func NewTemporalPostgres(dsn string, drop bool) (*RefreshRepoPostgres, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("gorm open db: %w", err)
	}

	if drop {
		err = db.Migrator().DropTable(&temporalRecord{})
		if err != nil {
			return nil, fmt.Errorf("gorm drop table: %w", err)
		}
	}

	err = db.AutoMigrate(&temporalRecord{})
	if err != nil {
		return nil, fmt.Errorf("gorm migrate table: %w", err)
	}

	return &RefreshRepoPostgres{
		db: db,
	}, nil
}

func (repo *RefreshRepoPostgres) NodesAt(height int32) ([]string, error) {

	var names []string

	err := repo.db.Model(&temporalRecord{}).
		Where(`? = ANY (heights)`, height).
		Pluck(`name`, &names).Error
	if err != nil {
		return nil, fmt.Errorf("gorm pluck: %w", err)
	}

	return names, nil
}

func (repo *RefreshRepoPostgres) SetNodeAt(name string, height int32) error {

	record := temporalRecord{Name: []byte(name)}

	err := repo.db.First(&record).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("gorm find: %w", err)
	}

	// Return early if it's already been set.
	for _, ht := range record.Heights {
		if ht == int64(height) {
			return nil
		}
	}
	record.Heights = append(record.Heights, int64(height))

	err = repo.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&record).Error
	if err != nil {
		return fmt.Errorf("gorm update: %w", err)
	}

	return nil
}

func (repo *RefreshRepoPostgres) Close() error {

	db, err := repo.db.DB()
	if err != nil {
		return fmt.Errorf("gorm get db: %w", err)
	}

	err = db.Close()
	if err != nil {
		return fmt.Errorf("gorm close db: %w", err)
	}

	return nil
}
