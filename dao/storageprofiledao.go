package dao

import (
	"github.com/skyrings/skyring-common/models"
)

type StorageProfileInterface interface {
	StorageProfile(ctxt string, name string) (sProfile models.StorageProfile, e error)
	StorageProfiles(ctxt string, query interface{}, ops models.QueryOps) (sProfiles []models.StorageProfile, e error)
	SaveStorageProfile(ctxt string, s models.StorageProfile) error
	DeleteStorageProfile(ctxt string, name string) error
	InitStorageProfile(ctxt string) error
}
