package audio

import (
	"encoding/json"
	"io/ioutil"
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
	dir  string
	data *presetData
}

func newPresetManager(dir string) *presetManager {
	return &presetManager{
		dir: dir,
	}
}

func (pm *presetManager) getList() ([]*presetMeta, error) {
	if pm.data == nil {
		pm.loadData()
	}
	return pm.data.list, nil
}
func (pm *presetManager) applyToParams(name string, target *params) error {
	path := pm.dir + "/" + name + ".json"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	target.applyJSON(bytes)
	return nil
}
func (pm *presetManager) loadData() error {
	path := pm.dir + "/_list.json"
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
