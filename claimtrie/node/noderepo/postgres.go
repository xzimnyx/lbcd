package noderepo

import (
	"fmt"

	"github.com/btcsuite/btcd/claimtrie/change"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Postgres struct {
	db *gorm.DB
}

type nodeRecord struct {
	ID     uint  `gorm:"primarykey;index:,type:brin"`
	Type   int   `gorm:"index"`
	Height int32 `gorm:"index:,type:brin"`

	Name     []byte `gorm:"index,type:hash"`
	ClaimID  string
	OutPoint string
	Amount   int64
	Value    []byte
}

func NewPostgres(dsn string, drop bool) (*Postgres, error) {

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("gotm open db: %w", err)
	}

	if drop {
		err = db.Migrator().DropTable(&nodeRecord{})
		if err != nil {
			return nil, fmt.Errorf("gorm drop table: %w", err)
		}
	}

	err = db.AutoMigrate(&nodeRecord{})
	if err != nil {
		return nil, fmt.Errorf("gorm migrate table: %w", err)
	}

	return &Postgres{db: db}, nil
}

func (repo *Postgres) Save(changes []change.Change) error {

	records := make([]nodeRecord, 0, len(changes))
	for _, chg := range changes {
		record := nodeRecord{
			Type:     int(chg.Type),
			Height:   chg.Height,
			Name:     chg.Name,
			ClaimID:  chg.ClaimID,
			OutPoint: chg.OutPoint,
			Amount:   chg.Amount,
			Value:    chg.Value,
		}
		records = append(records, record)
	}

	err := repo.db.Create(&records).Error
	if err != nil {
		return fmt.Errorf("gorm create: %w", err)
	}

	return nil
}

func (repo *Postgres) LoadByNameUpToHeight(name string, height int32) ([]change.Change, error) {

	var records []nodeRecord

	err := repo.db.
		Where("name = ? AND height <= ?", name, height).
		Order("id ASC").
		Find(&records).Error

	if err != nil {
		return nil, fmt.Errorf("gorm find: %w", err)
	}

	changes := make([]change.Change, 0, len(records))
	for _, record := range records {
		chg := change.Change{
			Type:     change.ChangeType(record.Type),
			Height:   record.Height,
			Name:     record.Name,
			ClaimID:  record.ClaimID,
			OutPoint: record.OutPoint,
			Amount:   record.Amount,
			Value:    record.Value,
		}

		changes = append(changes, chg)
	}

	return changes, nil
}

func (repo *Postgres) Close() error {

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
