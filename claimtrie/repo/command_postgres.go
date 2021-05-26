// Copyright (c) 2021 - LBRY Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package repo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/claimtrie/change"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ChangeRepoPostgres struct {
	db *gorm.DB

	changes []change.Change
	last    time.Time

	rows *sql.Rows
}

func NewChangeRepoPostgres(dsn string, drop bool) (*ChangeRepoPostgres, error) {

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if drop {
		db.Migrator().DropTable(&change.Change{})
	}
	db.AutoMigrate(&change.Change{})

	return &ChangeRepoPostgres{db: db}, nil
}

func (repo *ChangeRepoPostgres) Save(chg change.Change) error {

	repo.changes = append(repo.changes, chg)
	if len(repo.changes) > 100 || time.Since(repo.last) > time.Millisecond*200 {
		err := repo.db.Create(&repo.changes).Error
		if err != nil {
			return fmt.Errorf("gorm create: %w", err)
		}
		repo.last = time.Now()
		repo.changes = repo.changes[:0]
	}

	return nil
}

func (repo *ChangeRepoPostgres) Load() (change.Change, error) {

	var chg change.Change

	if repo.rows == nil {
		rows, err := repo.db.Model(&change.Change{}).Order("id ASC").Rows()
		if err != nil {
			return chg, fmt.Errorf("gorm rows: %w", err)
		}
		repo.rows = rows
	}

	if repo.rows.Next() {
		err := repo.db.ScanRows(repo.rows, &chg)
		if err != nil {
			return chg, fmt.Errorf("gorm scan rows: %w", err)
		}
	}

	return chg, nil
}

func (repo *ChangeRepoPostgres) Close() error {

	db, err := repo.db.DB()
	if err != nil {
		return fmt.Errorf("gorm db: %w", err)
	}

	return db.Close()
}
