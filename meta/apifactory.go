package meta

import (
	api "go-filecoin-storage-helper/lib/nodeapi"
	fapi "go-filecoin-storage-helper/lib/nodeapi/filecoin"
	lapi "go-filecoin-storage-helper/lib/nodeapi/lotus"

	"golang.org/x/xerrors"
)

func ApiFactory(vers string) (api.API, error) {
	var a api.API
	switch vers {
	case "lotus":
		a = lapi.NewLotusAPI()
	case "fil":
		a = fapi.NewGoFileAPI()
	default:
		return nil, xerrors.New("vers param err")
	}

	return a, nil
}
