package aerospike

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type aerospikeClient interface {
	Truncate(set string) error
	InsertBinMap(set, key string, binMap map[string]interface{}) error
}

type LoaderAerospike struct {
	client   aerospikeClient
	location string
	debug    bool
}

type binMap map[string]interface{}
type set map[string]binMap

type fixture struct {
	Inherits  []string
	Sets      yaml.MapSlice
	Templates yaml.MapSlice
}

type loadedSet struct {
	name string
	data set
}
type loadContext struct {
	files          []string
	sets           []loadedSet
	refsDefinition set
}

func New(client aerospikeClient, location string, debug bool) *LoaderAerospike {
	return &LoaderAerospike{
		client:   client,
		location: location,
		debug:    debug,
	}
}

func (l *LoaderAerospike) Load(names []string) error {
	ctx := loadContext{
		refsDefinition: make(set),
	}

	// Gather data from files.
	for _, name := range names {
		err := l.loadFile(name, &ctx)
		if err != nil {
			return fmt.Errorf("unable to load fixture %s: %s", name, err.Error())
		}
	}
	return l.loadSets(&ctx)
}

func (l *LoaderAerospike) loadFile(name string, ctx *loadContext) error {
	candidates := []string{
		l.location + "/" + name,
		l.location + "/" + name + ".yml",
		l.location + "/" + name + ".yaml",
	}
	var err error
	var file string
	for _, candidate := range candidates {
		if _, err = os.Stat(candidate); err == nil {
			file = candidate
			break
		}
	}
	if err != nil {
		return err
	}
	// skip previously loaded files
	if inArray(file, ctx.files) {
		return nil
	}
	if l.debug {
		fmt.Println("Loading", file)
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	ctx.files = append(ctx.files, file)
	return l.loadYml(data, ctx)
}

func (l *LoaderAerospike) loadYml(data []byte, ctx *loadContext) error {
	// read yml into struct
	var loadedFixture fixture
	if err := yaml.Unmarshal(data, &loadedFixture); err != nil {
		return err
	}

	// load inherits
	for _, inheritFile := range loadedFixture.Inherits {
		if err := l.loadFile(inheritFile, ctx); err != nil {
			return err
		}
	}

	// loadedFixture.templates
	// yaml.MapSlice{
	//    string => yaml.MapSlice{  --- template name
	//        string => interface{} --- bin name: value
	//    }
	// }
	for _, template := range loadedFixture.Templates {
		name := template.Key.(string)
		if _, ok := ctx.refsDefinition[name]; ok {
			return fmt.Errorf("unable to load template %s: duplicating ref name", name)
		}

		binMap, err := binMapFromYaml(template)
		if err != nil {
			return err
		}

		if base, ok := binMap["$extend"]; ok {
			baseName := base.(string)
			baseBinMap, err := l.resolveReference(ctx.refsDefinition, baseName)
			if err != nil {
				return err
			}
			for k, v := range binMap {
				baseBinMap[k] = v
			}
			binMap = baseBinMap
		}
		ctx.refsDefinition[name] = binMap
		if l.debug {
			marshalled, _ := json.Marshal(binMap)
			fmt.Printf("Populating ref %s as %s from template\n", name, string(marshalled))
		}
	}

	// loadedFixture.sets
	// yaml.MapSlice{
	//    string => yaml.MapSlice{      --- set name
	//        string => yaml.MapSlice{  --- key name
	//            string => interface{} --- bin name: value
	//        }
	//    }
	// }
	for _, yamlSet := range loadedFixture.Sets {
		set, err := setFromYaml(yamlSet)
		if err != nil {
			return err
		}
		lt := loadedSet{
			name: yamlSet.Key.(string),
			data: set,
		}
		ctx.sets = append(ctx.sets, lt)
	}
	return nil
}

func setFromYaml(mapItem yaml.MapItem) (set, error) {
	entries, ok := mapItem.Value.(yaml.MapSlice)
	if !ok {
		return nil, errors.New("expected map/array as set")
	}

	set := make(set, len(entries))
	for _, e := range entries {
		key := e.Key.(string)
		binmap, err := binMapFromYaml(e)
		if err != nil {
			return nil, err
		}
		set[key] = binmap
	}

	return set, nil
}

func binMapFromYaml(mapItem yaml.MapItem) (binMap, error) {
	bins, ok := mapItem.Value.(yaml.MapSlice)
	if !ok {
		return nil, errors.New("expected map/array as binmap")
	}

	binmap := make(binMap, len(bins))
	for j := range bins {
		binmap[bins[j].Key.(string)] = bins[j].Value
	}

	return binmap, nil
}

func (l *LoaderAerospike) loadSets(ctx *loadContext) error {
	// truncate first
	truncatedSets := make(map[string]bool)
	for _, s := range ctx.sets {
		if _, ok := truncatedSets[s.name]; ok {
			// already truncated
			continue
		}
		if err := l.truncateSet(s.name); err != nil {
			return err
		}
		truncatedSets[s.name] = true
	}

	// then load data
	for _, s := range ctx.sets {
		if len(s.data) == 0 {
			continue
		}
		if err := l.loadSet(ctx, s); err != nil {
			return fmt.Errorf("failed to load set '%s' because:\n%s", s.name, err)
		}
	}

	return nil
}

// truncateTable truncates table
func (l *LoaderAerospike) truncateSet(name string) error {
	return l.client.Truncate(name)
}

func (l *LoaderAerospike) loadSet(ctx *loadContext, set loadedSet) error {
	// $extend keyword allows, to import values from a named row
	for key, binMap := range set.data {
		if base, ok := binMap["$extend"]; ok {
			baseName := base.(string)
			baseBinMap, err := l.resolveReference(ctx.refsDefinition, baseName)
			if err != nil {
				return err
			}
			for k, v := range binMap {
				baseBinMap[k] = v
			}
			set.data[key] = baseBinMap
		}
	}

	for key, binmap := range set.data {
		err := l.client.InsertBinMap(set.name, key, binmap)
		if err != nil {
			return err
		}
	}

	return nil
}

// resolveReference finds previously stored reference by its name
func (l *LoaderAerospike) resolveReference(refs set, refName string) (binMap, error) {
	target, ok := refs[refName]
	if !ok {
		return nil, fmt.Errorf("undefined reference %s", refName)
	}
	// make a copy of referencing data to prevent spoiling the source
	// by the way removing $-records from base row
	targetCopy := make(binMap, len(target))
	for k, v := range target {
		if len(k) == 0 || k[0] != '$' {
			targetCopy[k] = v
		}
	}
	return targetCopy, nil
}

// inArray checks whether the needle is present in haystack slice
func inArray(needle string, haystack []string) bool {
	for _, e := range haystack {
		if needle == e {
			return true
		}
	}
	return false
}
