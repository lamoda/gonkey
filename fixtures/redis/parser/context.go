package parser

type Context struct {
	keyRefs  map[string]Keys
	hashRefs map[string]HashRecordValue
	setRefs  map[string]SetRecordValue
	listRefs map[string]ListRecordValue
	zsetRefs map[string]ZSetRecordValue
}

func NewContext() *Context {
	return &Context{
		keyRefs:  make(map[string]Keys),
		hashRefs: make(map[string]HashRecordValue),
		setRefs:  make(map[string]SetRecordValue),
		listRefs: make(map[string]ListRecordValue),
		zsetRefs: make(map[string]ZSetRecordValue),
	}
}
