/*
 * Copyright © 2019 One Concern
 *
 */

package core

import (
	"github.com/oneconcern/datamon/pkg/context"
	"github.com/oneconcern/datamon/pkg/storage"
)

func GetRepoStore(stores context.Stores) storage.Store {
	return getMetaStore(stores)
}
