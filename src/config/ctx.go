package config

import dbModel "github.com/ypMarkJo/wmLight/src/db/model"

type Ctx struct {
	Db *dbModel.DBConnection
}

var (
	AppCtx = &Ctx{
		Db: &dbModel.DBConnection{},
	}
)
