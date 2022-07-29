package parser

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestRedisFixtureParser_Load(t *testing.T) {
	type args struct {
		fixtures []string
	}

	type want struct {
		fixtures []*Fixture
		ctx      *context
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "test basic",
			args: args{
				fixtures: []string{"redis"},
			},
			want: want{
				fixtures: []*Fixture{
					{
						Databases: map[int]Database{
							1: {
								Keys: &Keys{
									Values: map[string]*KeyValue{
										"key1": {
											Value: Str("value1"),
										},
										"key2": {
											Value:      Str("value2"),
											Expiration: time.Second * 10,
										},
									},
								},
								Sets: &Sets{
									Values: map[string]*SetRecordValue{
										"set1": {
											Expiration: time.Second * 10,
											Values: []*SetValue{
												{Value: Str("a")},
												{Value: Str("b")},
												{Value: Str("c")},
											},
										},
										"set3": {
											Expiration: time.Second * 5,
											Values: []*SetValue{
												{Value: Str("x")},
												{Value: Str("y")},
											},
										},
									},
								},
								Hashes: &Hashes{
									Values: map[string]*HashRecordValue{
										"map1": {
											Values: []*HashValue{
												{Value: Int(1), Key: Str("a")},
												{Value: Int(2), Key: Str("b")},
											},
										},
										"map2": {
											Values: []*HashValue{
												{Value: Int(3), Key: Str("c")},
												{Value: Int(4), Key: Str("d")},
											},
										},
									},
								},
								Lists: &Lists{
									Values: map[string]*ListRecordValue{
										"list1": {
											Values: []*ListValue{
												{Value: Int(1)},
												{Value: Str("2")},
												{Value: Str("a")},
											},
										},
									},
								},
								ZSets: &ZSets{
									Values: map[string]*ZSetRecordValue{
										"zset1": {
											Values: []*ZSetValue{
												{Value: Int(1), Score: 1.1},
												{Value: Str("2"), Score: 5.6},
											},
										},
									},
								},
							},
							2: {
								Keys: &Keys{
									Values: map[string]*KeyValue{
										"key3": {
											Value: Str("value3"),
										},
										"key4": {
											Value:      Str("value4"),
											Expiration: time.Second * 5,
										},
									},
								},
								Sets: &Sets{
									Values: map[string]*SetRecordValue{
										"set2": {
											Expiration: time.Second * 5,
											Values: []*SetValue{
												{Value: Str("d")},
												{Value: Str("e")},
												{Value: Str("f")},
											},
										},
									},
								},
								Hashes: &Hashes{
									Values: map[string]*HashRecordValue{
										"map3": {
											Values: []*HashValue{
												{Value: Int(3), Key: Str("c")},
												{Value: Int(4), Key: Str("d")},
											},
										},
										"map4": {
											Values: []*HashValue{
												{Value: Int(10), Key: Str("e")},
												{Value: Int(11), Key: Str("f")},
											},
										},
									},
								},
							},
						},
					},
				},
				ctx: &context{
					keyRefs:  map[string]Keys{},
					setRefs:  map[string]SetRecordValue{},
					hashRefs: map[string]HashRecordValue{},
					listRefs: map[string]ListRecordValue{},
					zsetRefs: map[string]ZSetRecordValue{},
				},
			},
		},
		{
			name: "extend",
			args: args{
				fixtures: []string{"redis_extend"},
			},
			want: want{
				fixtures: []*Fixture{
					{
						Templates: Templates{
							Keys: []*Keys{
								{
									Name: "parentKeys",
									Values: map[string]*KeyValue{
										"a": {Value: Int(1)},
										"b": {Value: Int(2)},
									},
								},
								{
									Name:   "childKeys",
									Extend: "parentKeys",
									Values: map[string]*KeyValue{
										"a": {Value: Int(1)},
										"b": {Value: Int(2)},
										"c": {Value: Int(3)},
										"d": {Value: Int(4)},
									},
								},
							},
							Sets: []*SetRecordValue{
								{
									Name:       "parentSet",
									Expiration: time.Second * 10,
									Values: []*SetValue{
										{Value: Str("a")},
										{Value: Str("b")},
									},
								},
								{
									Name:       "childSet",
									Extend:     "parentSet",
									Expiration: time.Second * 10,
									Values: []*SetValue{
										{Value: Str("a")},
										{Value: Str("b")},
										{Value: Str("c")},
									},
								},
							},
							Hashes: []*HashRecordValue{
								{
									Name: "parentMap",
									Values: []*HashValue{
										{Value: Int(1), Key: Str("a1")},
										{Value: Int(2), Key: Str("b1")},
									},
								},
								{
									Name:   "childMap",
									Extend: "parentMap",
									Values: []*HashValue{
										{Value: Int(1), Key: Str("a1")},
										{Value: Int(2), Key: Str("b1")},
										{Value: Int(3), Key: Str("c1")},
									},
								},
							},
							Lists: []*ListRecordValue{
								{
									Name: "parentList",
									Values: []*ListValue{
										{Value: Int(1)},
										{Value: Int(2)},
									},
								},
								{
									Name: "childList",
									Extend: "parentList",
									Values: []*ListValue{
										{Value: Int(1)},
										{Value: Int(2)},
										{Value: Int(3)},
										{Value: Int(4)},
									},
								},
							},
							ZSets: []*ZSetRecordValue{
								{
									Name: "parentZSet",
									Values: []*ZSetValue{
										{Value: Int(1), Score: 1.2},
										{Value: Int(2), Score: 3.4},
									},
								},
								{
									Name: "childZSet",
									Extend: "parentZSet",
									Values: []*ZSetValue{
										{Value: Int(1), Score: 1.2},
										{Value: Int(2), Score: 3.4},
										{Value: Int(3), Score: 5.6},
										{Value: Int(4), Score: 7.8},
									},
								},
							},
						},
						Databases: map[int]Database{
							1: {
								Keys: &Keys{
									Extend: "childKeys",
									Values: map[string]*KeyValue{
										"a": {Value: Int(1)},
										"b": {Value: Int(2)},
										"c": {Value: Int(3)},
										"d": {Value: Int(4)},
										"key1": {
											Value: Str("value1"),
										},
										"key2": {
											Value:      Str("value2"),
											Expiration: time.Second * 10,
										},
									},
								},
								Sets: &Sets{
									Values: map[string]*SetRecordValue{
										"set1": {
											Extend:     "childSet",
											Expiration: time.Second * 10,
											Values: []*SetValue{
												{Value: Str("a")},
												{Value: Str("b")},
												{Value: Str("c")},
												{Value: Str("d")},
											},
										},
										"set2": {
											Expiration: time.Second * 10,
											Values: []*SetValue{
												{Value: Str("x")},
												{Value: Str("y")},
											},
										},
									},
								},
								Hashes: &Hashes{
									Values: map[string]*HashRecordValue{
										"map1": {
											Name:   "baseMap",
											Extend: "childMap",
											Values: []*HashValue{
												{Value: Int(1), Key: Str("a1")},
												{Value: Int(2), Key: Str("b1")},
												{Value: Int(3), Key: Str("c1")},
												{Value: Int(1), Key: Str("a")},
												{Value: Int(2), Key: Str("b")},
											},
										},
										"map2": {
											Values: []*HashValue{
												{Value: Int(3), Key: Str("c")},
												{Value: Int(4), Key: Str("d")},
											},
										},
									},
								},
								Lists: &Lists{
									Values: map[string]*ListRecordValue{
										"list1": {
											Name:   "list1",
											Extend: "childList",
											Values: []*ListValue{
												{Value: Int(1)},
												{Value: Int(2)},
												{Value: Int(3)},
												{Value: Int(4)},
												{Value: Int(10)},
												{Value: Int(11)},
											},
										},
									},
								},
								ZSets: &ZSets{
									Values: map[string]*ZSetRecordValue{
										"zset1": {
											Name:   "zset1",
											Extend: "childZSet",
											Values: []*ZSetValue{
												{Value: Int(1), Score: 1.2},
												{Value: Int(2), Score: 3.4},
												{Value: Int(3), Score: 5.6},
												{Value: Int(4), Score: 7.8},
												{Value: Int(5), Score: 10.1},
											},
										},
									},
								},
							},
						},
					},
				},
				ctx: &context{
					keyRefs: map[string]Keys{
						"parentKeys": {
							Values: map[string]*KeyValue{
								"a": {Value: Int(1)},
								"b": {Value: Int(2)},
							},
						},
						"childKeys": {
							Values: map[string]*KeyValue{
								"a": {Value: Int(1)},
								"b": {Value: Int(2)},
								"c": {Value: Int(3)},
								"d": {Value: Int(4)},
							},
						},
					},
					setRefs: map[string]SetRecordValue{
						"parentSet": {
							Expiration: time.Second * 10,
							Values: []*SetValue{
								{Value: Str("a")},
								{Value: Str("b")},
							},
						},
						"childSet": {
							Expiration: time.Second * 10,
							Values: []*SetValue{
								{Value: Str("a")},
								{Value: Str("b")},
								{Value: Str("c")},
							},
						},
					},
					hashRefs: map[string]HashRecordValue{
						"baseMap": {
							Values: []*HashValue{
								{Value: Int(1), Key: Str("a1")},
								{Value: Int(2), Key: Str("b1")},
								{Value: Int(3), Key: Str("c1")},
								{Value: Int(1), Key: Str("a")},
								{Value: Int(2), Key: Str("b")},
							},
						},
						"parentMap": {
							Values: []*HashValue{
								{Value: Int(1), Key: Str("a1")},
								{Value: Int(2), Key: Str("b1")},
							},
						},
						"childMap": {
							Values: []*HashValue{
								{Value: Int(1), Key: Str("a1")},
								{Value: Int(2), Key: Str("b1")},
								{Value: Int(3), Key: Str("c1")},
							},
						},
					},
					listRefs: map[string]ListRecordValue{
						"parentList": {
							Values: []*ListValue{
								{Value: Int(1)},
								{Value: Int(2)},
							},
						},
						"childList": {
							Values: []*ListValue{
								{Value: Int(1)},
								{Value: Int(2)},
								{Value: Int(3)},
								{Value: Int(4)},
							},
						},
						"list1": {
							Values: []*ListValue{
								{Value: Int(1)},
								{Value: Int(2)},
								{Value: Int(3)},
								{Value: Int(4)},
								{Value: Int(10)},
								{Value: Int(11)},
							},
						},
					},
					zsetRefs: map[string]ZSetRecordValue{
						"parentZSet": {
							Values: []*ZSetValue{
								{Value: Int(1), Score: 1.2},
								{Value: Int(2), Score: 3.4},
							},
						},
						"childZSet": {
							Values: []*ZSetValue{
								{Value: Int(1), Score: 1.2},
								{Value: Int(2), Score: 3.4},
								{Value: Int(3), Score: 5.6},
								{Value: Int(4), Score: 7.8},
							},
						},
						"zset1": {
							Values: []*ZSetValue{
								{Value: Int(1), Score: 1.2},
								{Value: Int(2), Score: 3.4},
								{Value: Int(3), Score: 5.6},
								{Value: Int(4), Score: 7.8},
								{Value: Int(5), Score: 10.1},
							},
						},
					},
				},
			},
		},
		{
			name: "inherits",
			args: args{
				fixtures: []string{"redis_inherits"},
			},
			want: want{
				fixtures: []*Fixture{
					{
						Inherits: []string{"redis_extend"},
						Databases: map[int]Database{
							1: {
								Keys: &Keys{
									Extend: "childKeys",
									Values: map[string]*KeyValue{
										"a":    {Value: Int(1)},
										"b":    {Value: Int(2)},
										"c":    {Value: Int(3)},
										"d":    {Value: Int(4)},
										"key1": {Value: Str("value1")},
										"key2": {Value: Str("value2"), Expiration: time.Second * 10},
									},
								},
								Sets: &Sets{
									Values: map[string]*SetRecordValue{
										"set1": {
											Extend:     "childSet",
											Expiration: time.Second * 10,
											Values: []*SetValue{
												{Value: Str("a")},
												{Value: Str("b")},
												{Value: Str("c")},
											},
										},
									},
								},
								Hashes: &Hashes{
									Values: map[string]*HashRecordValue{
										"map1": {
											Extend: "baseMap",
											Values: []*HashValue{
												{Value: Int(1), Key: Str("a1")},
												{Value: Int(2), Key: Str("b1")},
												{Value: Int(3), Key: Str("c1")},
												{Value: Int(1), Key: Str("a")},
												{Value: Int(2), Key: Str("b")},
												{Value: Int(10), Key: Str("x")},
												{Value: Int(11), Key: Str("y")},
											},
										},
										"map2": {
											Extend: "childMap",
											Values: []*HashValue{
												{Value: Int(1), Key: Str("a1")},
												{Value: Int(2), Key: Str("b1")},
												{Value: Int(3), Key: Str("c1")},
												{Value: Int(500), Key: Str("t")},
												{Value: Int(1000), Key: Str("j")},
											},
										},
									},
								},
								Lists: &Lists{
									Values: map[string]*ListRecordValue{
										"list2": {
											Extend: "list1",
											Values: []*ListValue{
												{Value: Int(1)},
												{Value: Int(2)},
												{Value: Int(3)},
												{Value: Int(4)},
												{Value: Int(10)},
												{Value: Int(11)},
												{Value: Int(100)},
											},
										},
										"list3": {
											Extend: "childList",
											Values: []*ListValue{
												{Value: Int(1)},
												{Value: Int(2)},
												{Value: Int(3)},
												{Value: Int(4)},
												{Value: Int(200)},
											},
										},
									},
								},
								ZSets: &ZSets{
									Values: map[string]*ZSetRecordValue{
										"zset2": {
											Extend: "zset1",
											Values: []*ZSetValue{
												{Value: Int(1), Score: 1.2},
												{Value: Int(2), Score: 3.4},
												{Value: Int(3), Score: 5.6},
												{Value: Int(4), Score: 7.8},
												{Value: Int(5), Score: 10.1},
												{Value: Int(100), Score: 100.1},
											},
										},
										"zset3": {
											Extend: "childZSet",
											Values: []*ZSetValue{
												{Value: Int(1), Score: 1.2},
												{Value: Int(2), Score: 3.4},
												{Value: Int(3), Score: 5.6},
												{Value: Int(4), Score: 7.8},
												{Value: Int(200), Score: 200.2},
											},
										},
									},
								},
							},
						},
					},
				},
				ctx: &context{
					keyRefs: map[string]Keys{
						"parentKeys": {
							Values: map[string]*KeyValue{
								"a": {Value: Int(1)},
								"b": {Value: Int(2)},
							},
						},
						"childKeys": {
							Values: map[string]*KeyValue{
								"a": {Value: Int(1)},
								"b": {Value: Int(2)},
								"c": {Value: Int(3)},
								"d": {Value: Int(4)},
							},
						},
					},
					setRefs: map[string]SetRecordValue{
						"parentSet": {
							Expiration: time.Second * 10,
							Values: []*SetValue{
								{Value: Str("a")},
								{Value: Str("b")},
							},
						},
						"childSet": {
							Expiration: time.Second * 10,
							Values: []*SetValue{
								{Value: Str("a")},
								{Value: Str("b")},
								{Value: Str("c")},
							},
						},
					},
					hashRefs: map[string]HashRecordValue{
						"baseMap": {
							Values: []*HashValue{
								{Value: Int(1), Key: Str("a1")},
								{Value: Int(2), Key: Str("b1")},
								{Value: Int(3), Key: Str("c1")},
								{Value: Int(1), Key: Str("a")},
								{Value: Int(2), Key: Str("b")},
							},
						},
						"parentMap": {
							Values: []*HashValue{
								{Value: Int(1), Key: Str("a1")},
								{Value: Int(2), Key: Str("b1")},
							},
						},
						"childMap": {
							Values: []*HashValue{
								{Value: Int(1), Key: Str("a1")},
								{Value: Int(2), Key: Str("b1")},
								{Value: Int(3), Key: Str("c1")},
							},
						},
					},
					listRefs: map[string]ListRecordValue{
						"parentList": {
							Values: []*ListValue{
								{Value: Int(1)},
								{Value: Int(2)},
							},
						},
						"childList": {
							Values: []*ListValue{
								{Value: Int(1)},
								{Value: Int(2)},
								{Value: Int(3)},
								{Value: Int(4)},
							},
						},
						"list1": {
							Values: []*ListValue{
								{Value: Int(1)},
								{Value: Int(2)},
								{Value: Int(3)},
								{Value: Int(4)},
								{Value: Int(10)},
								{Value: Int(11)},
							},
						},
					},
					zsetRefs: map[string]ZSetRecordValue{
						"parentZSet": {
							Values: []*ZSetValue{
								{Value: Int(1), Score: 1.2},
								{Value: Int(2), Score: 3.4},
							},
						},
						"childZSet": {
							Values: []*ZSetValue{
								{Value: Int(1), Score: 1.2},
								{Value: Int(2), Score: 3.4},
								{Value: Int(3), Score: 5.6},
								{Value: Int(4), Score: 7.8},
							},
						},
						"zset1": {
							Values: []*ZSetValue{
								{Value: Int(1), Score: 1.2},
								{Value: Int(2), Score: 3.4},
								{Value: Int(3), Score: 5.6},
								{Value: Int(4), Score: 7.8},
								{Value: Int(5), Score: 10.1},
							},
						},
					},
				},
			},
		},
	}

	p := New([]string{"../../testdata"})

	// test parsing example file from README
	_, err := p.ParseFiles(NewContext(), []string{"redis_example"})
	if err != nil {
		t.Errorf("example file test error: %s", err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := NewContext()
			fixtures, err := p.ParseFiles(ctx, test.args.fixtures)
			if err != nil {
				t.Errorf("ParseFiles - unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(test.want.fixtures, fixtures); diff != "" {
				t.Errorf("ParseFiles - unexpected diff in fixtures: %s", diff)
			}
			if diff := cmp.Diff(test.want.ctx, ctx, cmp.AllowUnexported(context{})); diff != "" {
				t.Errorf("ParseFiles - unexpected diff in context: %s", diff)
			}
		})
	}
}
