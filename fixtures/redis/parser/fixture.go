package parser

import (
	"fmt"
	"time"
)

// Fixture is a representation of the test data, that is preloaded to a redis database before the test starts
/* Example (yaml):
```yaml
inherits:
	- parent_template
	- child_template
	- other_fixture
databases:
	1:
		keys:
			$name: keys1
			values:
				a:
					value: 1
					expiration: 10s
				b:
					value: 2
	2:
		keys:
			$name: keys2
			values:
				c:
					value: 3
					expiration: 10s
				d:
					value: 4
```
*/
type Fixture struct {
	Inherits  []string         `yaml:"inherits"`
	Templates Templates        `yaml:"templates"`
	Databases map[int]Database `yaml:"databases"`
}

type Templates struct {
	Keys   []*Keys            `yaml:"keys"`
	Hashes []*HashRecordValue `yaml:"hashes"`
	Sets   []*SetRecordValue  `yaml:"sets"`
	Lists  []*ListRecordValue `yaml:"lists"`
	ZSets  []*ZSetRecordValue `yaml:"zsets"`
}

// Database contains data to load into Redis database
type Database struct {
	Keys   *Keys   `yaml:"keys"`
	Hashes *Hashes `yaml:"hashes"`
	Sets   *Sets   `yaml:"sets"`
	Lists  *Lists  `yaml:"lists"`
	ZSets  *ZSets  `yaml:"zsets"`
}

const (
	TypeInt   = "int"
	TypeStr   = "str"
	TypeFloat = "float"
)

type Value struct {
	Type  string
	Value interface{}
}

func (obj *Value) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var internal interface{}
	if err := unmarshal(&internal); err != nil {
		return err
	}
	switch v := internal.(type) {
	case string:
		obj.Type = TypeStr
		obj.Value = v
	case int, int16, int32, int64:
		obj.Type = TypeInt
		obj.Value = v
	case float64:
		obj.Type = TypeFloat
		obj.Value = v
	default:
		return fmt.Errorf("unknown value type: %T", v)
	}

	return nil
}

func Str(value string) Value {
	return Value{Type: TypeStr, Value: value}
}

func Int(value int) Value {
	return Value{Type: TypeInt, Value: value}
}

func Float(value float64) Value {
	return Value{Type: TypeFloat, Value: value}
}

// Keys represent a collection of key/value pairs, that will be loaded into Redis database
type Keys struct {
	Name   string               `yaml:"$name"`
	Extend string               `yaml:"$extend"`
	Values map[string]*KeyValue `yaml:"values"`
}

// KeyValue represent a redis key/value pair
type KeyValue struct {
	Value      Value         `yaml:"value"`
	Expiration time.Duration `yaml:"expiration"`
}

// Hashes represent a collection of hash data structures, that will be loaded into Redis database
type Hashes struct {
	Values map[string]*HashRecordValue `yaml:"values"`
}

// HashRecordValue represent a single hash data structure
type HashRecordValue struct {
	Name       string        `yaml:"$name"`
	Extend     string        `yaml:"$extend"`
	Values     []*HashValue  `yaml:"values"`
	Expiration time.Duration `yaml:"expiration"`
}

type HashValue struct {
	Key   Value `yaml:"key"`
	Value Value `yaml:"value"`
}

// Sets represent a collection of set data structures, that will be loaded into Redis database
type Sets struct {
	Values map[string]*SetRecordValue `yaml:"values"`
}

// SetRecordValue represent a single set data structure
type SetRecordValue struct {
	Name       string        `yaml:"$name"`
	Extend     string        `yaml:"$extend"`
	Values     []*SetValue   `yaml:"values"`
	Expiration time.Duration `yaml:"expiration"`
}

// SetValue represent a set value object
type SetValue struct {
	Value Value `yaml:"value"`
}

// Lists represent a collection of Redis list data structures
type Lists struct {
	Values map[string]*ListRecordValue `yaml:"values"`
}

// ListRecordValue represent a single list data structure
type ListRecordValue struct {
	Name       string        `yaml:"$name"`
	Extend     string        `yaml:"$extend"`
	Values     []*ListValue  `yaml:"values"`
	Expiration time.Duration `yaml:"expiration"`
}

// ListValue represent a list value object
type ListValue struct {
	Value Value `yaml:"value"`
}

// ZSets represent a collection of Redis sorted set data structure
type ZSets struct {
	Values map[string]*ZSetRecordValue `yaml:"values"`
}

// ZSetRecordValue represent a single sorted set data structure
type ZSetRecordValue struct {
	Name       string        `yaml:"$name"`
	Extend     string        `yaml:"$extend"`
	Values     []*ZSetValue  `yaml:"values"`
	Expiration time.Duration `yaml:"expiration"`
}

// ZSetValue represent a zset value object
type ZSetValue struct {
	Value Value   `yaml:"value"`
	Score float64 `yaml:"score"`
}
