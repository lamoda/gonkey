package mongo

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoaderMongo_loadYml(t *testing.T) {
	type args struct {
		data []byte
		ctx  *loadContext
	}
	tests := []struct {
		name string
		args args
		want loadContext
	}{
		{
			name: "basic",
			args: args{
				data: loadTestData(t, "../testdata/mongo.yaml"),
				ctx: &loadContext{
					refsDefinition: make(documentsDict),
					refsInserted:   make(documentsDict),
				},
			},
			want: loadContext{
				refsDefinition: documentsDict{},
				refsInserted:   documentsDict{},
				collections: []loadedCollection{
					{
						name: collectionName{database: "public", name: "collection1"},
						documents: collection{
							{
								"field1": "value1",
								"field2": 1,
							},
							{
								"field1": "value2",
								"field2": 2,
								"field3": 2.569947773654566473,
							},
						},
					},
					{
						name: collectionName{database: "public", name: "collection2"},
						documents: collection{
							{
								"field4": false,
								"field5": nil,
								"field1": `"`,
							},
							{
								"field1": "'",
								"field5": []interface{}{1, "2"},
							},
						},
					},
				},
			},
		},
		{
			name: "database",
			args: args{
				data: loadTestData(t, "../testdata/mongo_database.yaml"),
				ctx: &loadContext{
					refsDefinition: make(documentsDict),
					refsInserted:   make(documentsDict),
				},
			},
			want: loadContext{
				refsDefinition: documentsDict{},
				refsInserted:   documentsDict{},
				collections: []loadedCollection{
					{
						name: collectionName{database: "database1", name: "collection1"},
						documents: collection{
							{
								"f1": "value1",
								"f2": "value2",
							},
						},
					},
					{
						name: collectionName{database: "database2", name: "collection2"},
						documents: collection{
							{
								"f1": "value3",
								"f2": "value4",
							},
						},
					},
					{
						name: collectionName{database: "public", name: "collection3"},
						documents: collection{
							{
								"f1": "value5",
								"f2": "value6",
							},
						},
					},
				},
			},
		},
		{
			name: "extend",
			args: args{
				data: loadTestData(t, "../testdata/mongo_extend.yaml"),
				ctx: &loadContext{
					refsDefinition: documentsDict{},
					refsInserted:   documentsDict{},
				},
			},
			want: loadContext{
				refsDefinition: documentsDict{
					"base_tmpl": {
						"field1": "value1",
					},
					"extended_tmpl": {
						"$extend": "base_tmpl",
						"field1":  "value1",
						"field2":  "value2",
					},
				},
				refsInserted: documentsDict{},
				collections: []loadedCollection{
					{
						name: collectionName{database: "public", name: "collection1"},
						documents: collection{
							{
								"$extend": "base_tmpl",
								"field1":  "overwritten",
							},
						},
					},
					{
						name: collectionName{database: "public", name: "collection2"},
						documents: collection{
							{
								"$extend": "extended_tmpl",
								"field2":  "overwritten",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &LoaderMongo{}
			if err := f.loadYml(tt.args.data, tt.args.ctx); err != nil {
				t.Errorf("LoaderMongo.loadYml() error = %v", err)
			}

			require.Equal(t, tt.want, *tt.args.ctx)
		})
	}
}

func loadTestData(t *testing.T, path string) []byte {
	yml, err := ioutil.ReadFile(path)
	if err != nil {
		t.Error("No " + path)
	}
	return yml
}
