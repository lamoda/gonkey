package mongo

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

type mongoClient interface {
	Truncate(database string, collection string) error
	InsertDocument(database string, collection string, document map[string]interface{}) error
}

type LoaderMongo struct {
	client   mongoClient
	location string
	debug    bool
}

type document map[string]interface{}

type collection []document

type documentsDict map[string]document

type fixture struct {
	Inherits    []string
	Collections yaml.MapSlice
	Templates   yaml.MapSlice
}

type loadedCollection struct {
	name      collectionName
	documents collection
}

type collectionName struct {
	name     string
	database string
}

func newCollectionName(source string) collectionName {
	parts := strings.SplitN(source, ".", 2)

	switch {
	case len(parts) == 1:
		parts = append(parts, parts[0])
		fallthrough
	case parts[0] == "":
		parts[0] = "public"
	}

	cn := collectionName{database: parts[0], name: parts[1]}
	return cn
}

func (t *collectionName) getFullName() string {
	return fmt.Sprintf("\"%s\".\"%s\"", t.database, t.name)
}

type loadContext struct {
	files          []string
	collections    []loadedCollection
	refsDefinition documentsDict
	refsInserted   documentsDict
}

func New(client mongoClient, location string, debug bool) *LoaderMongo {
	return &LoaderMongo{
		client:   client,
		location: location,
		debug:    debug,
	}
}

func (f *LoaderMongo) Load(names []string) error {
	ctx := loadContext{
		refsDefinition: make(documentsDict),
		refsInserted:   make(documentsDict),
	}

	// gather data from files
	for _, name := range names {
		if err := f.loadFile(name, &ctx); err != nil {
			return fmt.Errorf("unable to load fixture %s: %s", name, err.Error())
		}
	}

	return f.loadCollections(&ctx)
}

func (f *LoaderMongo) loadFile(name string, ctx *loadContext) error {
	candidates := []string{
		f.location + "/" + name,
		f.location + "/" + name + ".yml",
		f.location + "/" + name + ".yaml",
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

	if f.debug {
		fmt.Println("Loading", file)
	}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	ctx.files = append(ctx.files, file)
	return f.loadYml(data, ctx)
}

func (f *LoaderMongo) loadYml(data []byte, ctx *loadContext) error {
	// read yml into struct
	var loadedFixture fixture
	if err := yaml.Unmarshal(data, &loadedFixture); err != nil {
		return err
	}

	// load inherits
	for _, inheritFile := range loadedFixture.Inherits {
		if err := f.loadFile(inheritFile, ctx); err != nil {
			return err
		}
	}

	// loadedFixture.templates
	// yaml.MapSlice{
	//    string => yaml.MapSlice{	--- template name
	//        string => interface{}	--- field name: value
	//    }
	// }
	for _, template := range loadedFixture.Templates {
		name := template.Key.(string)
		if _, ok := ctx.refsDefinition[name]; ok {
			return fmt.Errorf("unable to load template %s: duplicating ref name", name)
		}

		fields := template.Value.(yaml.MapSlice)
		doc := make(document, len(fields))
		for _, field := range fields {
			key := field.Key.(string)
			value, _ := field.Value.(interface{})
			doc[key] = value
		}

		if base, ok := doc["$extend"]; ok {
			base := base.(string)
			baseDoc, err := f.resolveReference(ctx.refsDefinition, base)
			if err != nil {
				return err
			}
			for k, v := range doc {
				baseDoc[k] = v
			}
			doc = baseDoc
		}
		ctx.refsDefinition[name] = doc

		if f.debug {
			rowJson, _ := json.Marshal(doc)
			fmt.Printf("Populating ref %s as %s from template\n", name, string(rowJson))
		}
	}

	// loadedFixture.collections
	// yaml.MapSlice{
	//    string => []interface{		--- collection name
	//        yaml.MapSlice{			--- document
	//            string => interface{}	--- field name: value
	//        }
	//    }
	// }
	for _, sourceCollection := range loadedFixture.Collections {
		sourceDocuments, ok := sourceCollection.Value.([]interface{})
		if !ok {
			return errors.New("expected array at root level")
		}

		documents := make(collection, len(sourceDocuments))
		for i := range sourceDocuments {
			sourceFields := sourceDocuments[i].(yaml.MapSlice)
			fields := make(document, len(sourceFields))
			for j := range sourceFields {
				fields[sourceFields[j].Key.(string)] = sourceFields[j].Value
			}
			documents[i] = fields
		}

		lc := loadedCollection{
			name:      newCollectionName(sourceCollection.Key.(string)),
			documents: documents,
		}
		ctx.collections = append(ctx.collections, lc)
	}

	return nil
}

func (f *LoaderMongo) loadCollections(ctx *loadContext) error {
	// truncate first
	if err := f.truncateCollections(ctx.collections); err != nil {
		return err
	}

	// then load data
	for _, cl := range ctx.collections {
		if len(cl.documents) == 0 {
			continue
		}

		if err := f.loadCollection(ctx, cl); err != nil {
			return fmt.Errorf("failed to load collection '%s' because:\n%s", cl.name.getFullName(), err)
		}
	}

	return nil
}

// truncateCollections truncates collection
func (f *LoaderMongo) truncateCollections(collections []loadedCollection) error {
	truncatedCollections := make(map[string]bool)
	for _, cl := range collections {
		if _, ok := truncatedCollections[cl.name.getFullName()]; ok {
			// already truncated
			continue
		}

		if err := f.client.Truncate(cl.name.database, cl.name.name); err != nil {
			return err
		}

		truncatedCollections[cl.name.getFullName()] = true
	}

	return nil
}

func (f *LoaderMongo) loadCollection(ctx *loadContext, cl loadedCollection) error {
	// $extend keyword allows, to import values from a named row
	for i, doc := range cl.documents {
		if base, ok := doc["$extend"]; ok {
			baseName := base.(string)
			baseDoc, err := f.resolveReference(ctx.refsDefinition, baseName)
			if err != nil {
				return err
			}

			for k, v := range doc {
				baseDoc[k] = v
			}

			cl.documents[i] = baseDoc
		}
	}

	for _, doc := range cl.documents {
		if err := f.client.InsertDocument(cl.name.database, cl.name.name, doc); err != nil {
			return err
		}
	}

	return nil
}

// resolveReference finds previously stored reference by its name
func (f *LoaderMongo) resolveReference(refs documentsDict, refName string) (document, error) {
	target, ok := refs[refName]
	if !ok {
		return nil, fmt.Errorf("undefined reference %s", refName)
	}

	// make a copy of referencing data to prevent spoiling the source
	// by the way removing $-records from base row
	targetCopy := make(document, len(target))
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
