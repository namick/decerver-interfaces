package dapps

import (
	"encoding/json"
)

const (
	PACKAGE_FILE_NAME  = "package.json"
	INDEX_FILE_NAME    = "index.html"
	MODELS_FOLDER_NAME = "models"
)

// Structs that are mapped to the package file.
type (
	PackageFile struct {
		Name               string              `json:"name"`
		Id                 string              `json:"id"`
		Icon               string              `json:"app_icon"`
		Version            string              `json:"version"`
		Homepage           string              `json:"homepage"`
		Author             *Author             `json:"author"`
		Repository         *Repository         `json:"repository"`
		Bugs               *Bugs               `json:"bugs"`
		Licence            *Licence            `json:"licence"`
		ModuleDependencies []*ModuleDependency `json:"module_dependencies"`
	}

	Author struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}

	Repository struct {
		Type string `json:"type"`
		Url  string `json:"url"`
	}

	Bugs struct {
		Url string `json:"url"`
	}

	Licence struct {
		Type string `json:"type"`
		Url  string `json:"url"`
	}

	ModuleDependency struct {
		Name    string          `json:"name"`
		Version string          `json:"version"`
		Data    json.RawMessage `json:data`
	}

	MonkData struct {
		ChainId        string `json:"chain_id"`
		PeerServerAddress string `json:"peer_server_address"`
	}
)

func NewPackageFileFromJson(pfJson []byte) (*PackageFile, error) {
	pf := &PackageFile{}
	err := json.Unmarshal(pfJson, pf)
	if err != nil {
		return nil, err
	}
	return pf, nil
}
