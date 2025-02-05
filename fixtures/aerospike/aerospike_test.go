package aerospike

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoaderAerospike_loadYml(t *testing.T) {
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
				data: loadTestData(t, "../testdata/aerospike.yaml"),
				ctx: &loadContext{
					refsDefinition: make(set),
				},
			},
			want: loadContext{
				refsDefinition: set{},
				sets: []loadedSet{
					{
						name: "set1",
						data: set{
							"key1": {
								"bin1": "value1",
								"bin2": 1,
							},
							"key2": {
								"bin1": "value2",
								"bin2": 2,
								"bin3": 2.569947773654566473,
							},
						},
					},
					{
						name: "set2",
						data: set{
							"key1": {
								"bin1": `"`,
								"bin4": false,
								"bin5": nil,
							},
							"key2": {
								"bin1": "'",
								"bin5": []interface{}{1, "2"},
							},
						},
					},
				},
			},
		},
		{
			name: "extend",
			args: args{
				data: loadTestData(t, "../testdata/aerospike_extend.yaml"),
				ctx: &loadContext{
					refsDefinition: make(set),
				},
			},
			want: loadContext{
				sets: []loadedSet{
					{
						name: "set1",
						data: set{
							"key1": {
								"$extend": "base_tmpl",
								"bin1":    "overwritten",
							},
						},
					},
					{
						name: "set2",
						data: set{
							"key1": {
								"$extend": "extended_tmpl",
								"bin2":    "overwritten",
							},
						},
					},
				},
				refsDefinition: set{
					"base_tmpl": {
						"bin1": "value1",
					},
					"extended_tmpl": {
						"$extend": "base_tmpl",
						"bin1":    "value1",
						"bin2":    "value2",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &LoaderAerospike{}
			if err := f.loadYml(tt.args.data, tt.args.ctx); err != nil {
				t.Errorf("LoaderAerospike.loadYml() error = %v", err)
			}

			require.Equal(t, tt.want, *tt.args.ctx)
		})
	}
}

func loadTestData(t *testing.T, path string) []byte {
	aerospikeYaml, err := os.ReadFile(path)
	if err != nil {
		t.Error("No aerospike.yaml")
	}
	return aerospikeYaml
}
