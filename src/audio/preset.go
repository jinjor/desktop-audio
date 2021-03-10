package audio

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
)

type presetMetaJSON struct {
	Name string `json:"name"`
}
type presetMetaListJSON struct {
	Items []presetMetaJSON `json:"items"`
}
type presetMeta struct {
	name string
}
type presetData struct {
	list []*presetMeta
}
type presetManager struct {
	selected string
	dir      string
	data     *presetData
}

func newPresetManager(dir string) *presetManager {
	return &presetManager{
		dir: dir,
	}
}
func (pm *presetManager) _nameToJSONPath(name string) string {
	return pm.dir + "/" + url.PathEscape(name) + ".json"
}

// ----- Params ----- //

func (pm *presetManager) applyToParams(name string, target *params) error {
	path := pm._nameToJSONPath(name)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	target.applyJSON(bytes)
	pm.selected = name
	return nil
}
func (pm *presetManager) restoreLastParams(p *params) (bool, error) {
	return pm._loadParams("_tmp", p)
}
func (pm *presetManager) _loadParams(name string, p *params) (bool, error) {
	path := pm._nameToJSONPath(name)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return false, nil
	}
	p.applyJSON(bytes)
	return true, nil
}
func (pm *presetManager) saveTemporaryParams(p *params) error {
	return pm._saveParams("_tmp", p)
}
func (pm *presetManager) overrideParams(p *params) (bool, error) {
	return pm.saveParams(pm.selected, p)
}
func (pm *presetManager) saveParams(name string, p *params) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("empty name cannot be accepted")
	}
	if name[0] == '_' {
		return false, fmt.Errorf("names start with _ cannot be accepted")
	}
	err := pm._saveParams(name, p)
	if err != nil {
		return false, err
	}
	return pm._upsertList(name)
}
func (pm *presetManager) _saveParams(name string, p *params) error {
	os.MkdirAll(pm.dir, os.ModePerm)
	path := pm._nameToJSONPath(name)
	j := p.toJSON()
	return ioutil.WriteFile(path, j, 0666)
}
func (pm *presetManager) remove(name string) (bool, error) {
	if name == "" {
		return false, fmt.Errorf("empty name cannot be accepted")
	}
	if name[0] == '_' {
		return false, fmt.Errorf("names start with _ cannot be accepted")
	}
	err := pm._removeParams(name)
	if err != nil {
		return false, err
	}
	return pm._removeFromList(name)
}
func (pm *presetManager) _removeParams(name string) error {
	os.MkdirAll(pm.dir, os.ModePerm)
	path := pm._nameToJSONPath(name)
	return os.Remove(path)
}

// ----- List ----- //

func (pm *presetManager) existsInList(name string) (bool, error) {
	if err := pm._ensureList(); err != nil {
		return false, err
	}
	for _, meta := range pm.data.list {
		if meta.name == name {
			return true, nil
		}
	}
	return false, nil
}
func (pm *presetManager) getList() ([]*presetMeta, error) {
	if err := pm._ensureList(); err != nil {
		return nil, err
	}
	return pm.data.list, nil
}
func (pm *presetManager) _upsertList(name string) (bool, error) {
	for _, meta := range pm.data.list {
		if meta.name == name {
			return false, nil
		}
	}
	pm.data.list = append(pm.data.list, &presetMeta{name: name})
	return true, pm._saveList()
}
func (pm *presetManager) _removeFromList(name string) (bool, error) {
	found := false
	for i := len(pm.data.list) - 1; i >= 0; i-- {
		meta := pm.data.list[i]
		if meta.name == name {
			pm.data.list = append(pm.data.list[:i], pm.data.list[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return false, nil
	}
	return true, pm._saveList()
}
func (pm *presetManager) _saveList() error {
	os.MkdirAll(pm.dir, os.ModePerm)
	path := pm._nameToJSONPath("_list")
	j, err := pm._listToJSON()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, j, 0666)
}
func (pm *presetManager) _saveInitialList() error {
	os.MkdirAll(pm.dir, os.ModePerm)
	path := pm._nameToJSONPath("_list")
	return ioutil.WriteFile(path, []byte(`{"items":[]}`), 0666)
}
func (pm *presetManager) listToJSON() (json.RawMessage, error) {
	if err := pm._ensureList(); err != nil {
		return nil, err
	}
	return pm._listToJSON()
}
func (pm *presetManager) _listToJSON() (json.RawMessage, error) {
	listJSON := &presetMetaListJSON{
		Items: make([]presetMetaJSON, len(pm.data.list)),
	}
	for i, item := range pm.data.list {
		listJSON.Items[i] = presetMetaJSON{
			Name: item.name,
		}
	}
	return toRawMessage(&listJSON), nil
}
func (pm *presetManager) _ensureList() error {
	if pm.data == nil {
		return pm._loadList()
	}
	return nil
}
func (pm *presetManager) _loadList() error {
	path := pm._nameToJSONPath("_list")
	if _, err := os.Stat(path); err != nil {
		err := pm._saveInitialList()
		if err != nil {
			return err
		}
	}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	metaListJSON := &presetMetaListJSON{}
	err = json.Unmarshal(bytes, &metaListJSON)
	if err != nil {
		return err
	}
	if pm.data == nil {
		pm.data = &presetData{list: make([]*presetMeta, 0, 128)}
	}
	pm.data.list = pm.data.list[0:len(metaListJSON.Items)]
	for i, item := range metaListJSON.Items {
		pm.data.list[i] = &presetMeta{name: item.Name}
	}
	return nil
}
