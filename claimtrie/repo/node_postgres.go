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
	"fmt"

	"github.com/btcsuite/btcd/claimtrie/change"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type NodeRepoPostgres struct {
	db *gorm.DB
}

func NewNodeRepoPostgres(dsn string) (*NodeRepoPostgres, error) {

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	return &NodeRepoPostgres{db: db}, nil
}

func (repo *NodeRepoPostgres) Load(name string, height int32) ([]change.Change, error) {

	var changes []change.Change

	err := repo.db.
		Where("name = ? AND height <= ?", name, height).
		Order("id ASC").
		Find(&changes).Error
	if err != nil {
		return nil, fmt.Errorf("gorm find: %w", err)
	}

	return changes, nil
}

func (repo *NodeRepoPostgres) Close() error {

	db, err := repo.db.DB()
	if err != nil {
		return fmt.Errorf("gorm db: %w", err)
	}

	return db.Close()
}
