package parser

import (
    "errors"
    "fmt"
    "io/ioutil"

    "gopkg.in/yaml.v3"
)

type redisYamlParser struct {
    fileParser *fileParser
}

func (p *redisYamlParser) Copy(fileParser *fileParser) FixtureFileParser {
    cp := &(*p)
    cp.fileParser = fileParser
    return cp
}

func (p *redisYamlParser) extendKeys(ctx *context, child *Keys) error {
    if child.Extend == "" {
        return nil
    }
    parent, err := p.resolveKeyReference(ctx.keyRefs, child.Extend)
    if err != nil {
        return err
    }
    for k, v := range child.Values {
        parent.Values[k] = v
    }
    child.Values = parent.Values
    return nil
}

func (p *redisYamlParser) copyKeyRecord(src *Keys) Keys {
    keyRef := Keys{
        Values: make(map[string]*KeyValue, len(src.Values)),
    }
    for k, v := range src.Values {
        var valueCopy *KeyValue
        if v != nil {
            valueCopy = &(*v)
        }
        keyRef.Values[k] = valueCopy
    }
    return keyRef
}

func (p *redisYamlParser) extendSet(ctx *context, child *SetRecordValue) error {
    if child.Extend == "" {
        return nil
    }
    parent, err := p.resolveSetReference(ctx.setRefs, child.Extend)
    if err != nil {
        return err
    }
    var keys []interface{}
    parentValuesMapped := make(map[interface{}]*SetValue)
    for _, v := range parent.Values {
        parentValuesMapped[v] = v
        keys = append(keys, v)
    }
    for _, v := range child.Values {
        if _, ok := parentValuesMapped[v]; !ok {
            keys = append(keys, v)
        }
        parentValuesMapped[v] = v
    }
    setValues := make([]*SetValue, 0, len(parentValuesMapped))
    for _, k := range keys {
        setValues = append(setValues, parentValuesMapped[k])
    }
    child.Expiration = parent.Expiration
    child.Values = setValues
    return nil
}

func (p *redisYamlParser) copySetRecord(src *SetRecordValue) SetRecordValue{
    setRef := SetRecordValue{
        Expiration: src.Expiration,
        Values:     make([]*SetValue, 0, len(src.Values)),
    }
    for _, v := range src.Values {
        var valueCopy *SetValue
        if v != nil {
            valueCopy = &(*v)
        }
        setRef.Values = append(setRef.Values, valueCopy)
    }
    return setRef
}

func (p *redisYamlParser) extendHash(ctx *context, child *HashRecordValue) error {
    if child.Extend == "" {
        return nil
    }
    parent, err := p.resolveHashReference(ctx.hashRefs, child.Extend)
    if err != nil {
        return err
    }
    var keys []interface{}
    parentValuesMapped := make(map[interface{}]*HashValue)
    for _, v := range parent.Values {
        parentValuesMapped[v.Key] = v
        keys = append(keys, v.Key)
    }
    for _, v := range child.Values {
        if _, ok := parentValuesMapped[v.Key]; !ok {
            keys = append(keys, v.Key)
        }
        parentValuesMapped[v.Key] = v
    }
    hashValues := make([]*HashValue, 0, len(parentValuesMapped))
    for _, k := range keys {
        hashValues = append(hashValues, parentValuesMapped[k])
    }

    child.Expiration = parent.Expiration
    child.Values = hashValues
    return nil
}

func (p *redisYamlParser) copyHashRecord(src *HashRecordValue) HashRecordValue {
    cpy := HashRecordValue{
        Expiration: src.Expiration,
        Values:     make([]*HashValue, 0, len(src.Values)),
    }
    for _, v := range src.Values {
        var valueCopy *HashValue
        if v != nil {
            valueCopy = &(*v)
        }
        cpy.Values = append(cpy.Values, valueCopy)
    }
    return cpy
}

func (p *redisYamlParser) extendList(ctx *context, child *ListRecordValue) error {
    if child.Extend == "" {
        return nil
    }
    parent, err := p.resolveListReference(ctx.listRefs, child.Extend)
    if err != nil {
        return err
    }
    for _, v := range child.Values {
        parent.Values = append(parent.Values, v)
    }
    child.Expiration = parent.Expiration
    child.Values = parent.Values
    return nil
}

func (p *redisYamlParser) copyListRecord(src *ListRecordValue) ListRecordValue {
    ref := ListRecordValue{
        Expiration: src.Expiration,
        Values:     make([]*ListValue, 0, len(src.Values)),
    }
    for _, v := range src.Values {
        var valueCopy *ListValue
        if v != nil {
            valueCopy = &(*v)
        }
        ref.Values = append(ref.Values, valueCopy)
    }
    return ref
}

func (p *redisYamlParser) extendZSet(ctx *context, child *ZSetRecordValue) error {
    if child.Extend == "" {
        return nil
    }
    parent, err := p.resolveZSetReference(ctx.zsetRefs, child.Extend)
    if err != nil {
        return err
    }
    var keys []interface{}
    parentValuesMapped := make(map[interface{}]*ZSetValue)
    for _, v := range parent.Values {
        parentValuesMapped[v] = v
        keys = append(keys, v)
    }
    for _, v := range child.Values {
        if _, ok := parentValuesMapped[v]; !ok {
            keys = append(keys, v)
        }
        parentValuesMapped[v] = v
    }
    setValues := make([]*ZSetValue, 0, len(parentValuesMapped))
    for _, k := range keys {
        setValues = append(setValues, parentValuesMapped[k])
    }

    child.Expiration = parent.Expiration
    child.Values = setValues
    return nil
}

func (p *redisYamlParser) copyZSetRecord(src *ZSetRecordValue) ZSetRecordValue {
    ref := ZSetRecordValue{
        Expiration: src.Expiration,
        Values:     make([]*ZSetValue, 0, len(src.Values)),
    }
    for _, v := range src.Values {
        var valueCopy *ZSetValue
        if v != nil {
            valueCopy = &(*v)
        }
        ref.Values = append(ref.Values, valueCopy)
    }
    return ref
}

func (p *redisYamlParser) buildKeysTemplates(ctx *context, f Fixture) error {
    for _, tplData := range f.Templates.Keys {
        refName := tplData.Name
        if refName == "" {
            return errors.New("template $name is required")
        }
        if _, ok := ctx.keyRefs[refName]; ok {
            return fmt.Errorf("unable to load template %s: duplicating ref name", refName)
        }
        if err := p.extendKeys(ctx, tplData); err != nil {
            return err
        }
        ctx.keyRefs[refName] = p.copyKeyRecord(tplData)
    }
    return nil
}

func (p *redisYamlParser) buildSetTemplates(ctx *context, f Fixture) error {
    for _, tplData := range f.Templates.Sets {
        refName := tplData.Name
        if refName == "" {
            return errors.New("template $name is required")
        }
        if _, ok := ctx.setRefs[refName]; ok {
            return fmt.Errorf("unable to load template %s: duplicating ref name", refName)
        }
        if err := p.extendSet(ctx, tplData); err != nil {
            return err
        }
        ctx.setRefs[refName] = p.copySetRecord(tplData)
    }
    return nil
}

func (p *redisYamlParser) buildHashTemplates(ctx *context, f Fixture) error {
    for _, tplData := range f.Templates.Hashes {
        refName := tplData.Name
        if refName == "" {
            return errors.New("template $name is required")
        }
        if _, ok := ctx.hashRefs[refName]; ok {
            return fmt.Errorf("unable to load template %s: duplicating ref name", refName)
        }
        if err := p.extendHash(ctx, tplData); err != nil {
            return err
        }
        ctx.hashRefs[refName] = p.copyHashRecord(tplData)
    }
    return nil
}

func (p *redisYamlParser) buildListTemplates(ctx *context, f Fixture) error {
    for _, tplData := range f.Templates.Lists {
        refName := tplData.Name
        if refName == "" {
            return errors.New("template $name is required")
        }
        if _, ok := ctx.listRefs[refName]; ok {
            return fmt.Errorf("unable to load template %s: duplicating ref name", refName)
        }
        if err := p.extendList(ctx, tplData); err != nil {
            return err
        }
        ctx.listRefs[refName] = p.copyListRecord(tplData)
    }
    return nil
}

func (p *redisYamlParser) buildZSetTemplates(ctx *context, f Fixture) error {
    for _, tplData := range f.Templates.ZSets {
        refName := tplData.Name
        if refName == "" {
            return errors.New("template $name is required")
        }
        if _, ok := ctx.zsetRefs[refName]; ok {
            return fmt.Errorf("unable to load template %s: duplicating ref name", refName)
        }
        if err := p.extendZSet(ctx, tplData); err != nil {
            return err
        }
        ctx.zsetRefs[refName] = p.copyZSetRecord(tplData)
    }
    return nil
}

func (p *redisYamlParser) buildTemplate(ctx *context, f Fixture) error {
    if err := p.buildKeysTemplates(ctx, f); err != nil {
        return err
    }
    if err := p.buildSetTemplates(ctx, f); err != nil {
        return err
    }
    if err := p.buildHashTemplates(ctx, f); err != nil {
        return err
    }
    if err := p.buildListTemplates(ctx, f); err != nil {
        return err
    }
    if err := p.buildZSetTemplates(ctx, f); err != nil {
        return err
    }
    return nil
}

func (p *redisYamlParser) resolveKeyReference(refs map[string]Keys, refName string) (*Keys, error) {
    refTemplate, ok := refs[refName]
    if !ok {
        return nil, fmt.Errorf("ref not found: %s", refName)
    }
    cpy := p.copyKeyRecord(&refTemplate)
    return &cpy, nil
}

func (p *redisYamlParser) resolveSetReference(refs map[string]SetRecordValue, refName string) (*SetRecordValue, error) {
    refTemplate, ok := refs[refName]
    if !ok {
        return nil, fmt.Errorf("ref not found: %s", refName)
    }
    cpy := p.copySetRecord(&refTemplate)
    return &cpy, nil
}

func (p *redisYamlParser) resolveHashReference(refs map[string]HashRecordValue, refName string) (*HashRecordValue, error) {
    refTemplate, ok := refs[refName]
    if !ok {
        return nil, fmt.Errorf("ref not found: %s", refName)
    }
    cpy := p.copyHashRecord(&refTemplate)
    return &cpy, nil
}

func (p *redisYamlParser) resolveListReference(refs map[string]ListRecordValue, refName string) (*ListRecordValue, error) {
    refTemplate, ok := refs[refName]
    if !ok {
        return nil, fmt.Errorf("ref not found: %s", refName)
    }
    cpy := p.copyListRecord(&refTemplate)
    return &cpy, nil
}

func (p *redisYamlParser) resolveZSetReference(refs map[string]ZSetRecordValue, refName string) (*ZSetRecordValue, error) {
    refTemplate, ok := refs[refName]
    if !ok {
        return nil, fmt.Errorf("ref not found: %s", refName)
    }
    cpy := p.copyZSetRecord(&refTemplate)
    return &cpy, nil
}

func (p *redisYamlParser) buildKeys(ctx *context, data *Keys) error {
    if data == nil {
        return nil
    }
    if err := p.extendKeys(ctx, data); err != nil {
        return err
    }
    if data.Name != "" {
        ctx.keyRefs[data.Name] = p.copyKeyRecord(data)
    }
    return nil
}

func (p *redisYamlParser) buildSets(ctx *context, data *Sets) error {
    if data == nil {
        return nil
    }
    for _, v := range data.Values {
        if err := p.extendSet(ctx, v); err != nil {
            return fmt.Errorf("extend set error: %w", err)
        }
        if v.Name != "" {
            ctx.setRefs[v.Name] = p.copySetRecord(v)
        }
    }
    return nil
}

func (p *redisYamlParser) buildMaps(ctx *context, data *Hashes) error {
    if data == nil {
        return nil
    }
    for _, v := range data.Values {
        if err := p.extendHash(ctx, v); err != nil {
            return fmt.Errorf("extend hash error: %w", err)
        }
        if v.Name != "" {
            ctx.hashRefs[v.Name] = p.copyHashRecord(v)
        }
    }
    return nil
}

func (p *redisYamlParser) buildLists(ctx *context, data *Lists) error {
    if data == nil {
        return nil
    }
    for _, v := range data.Values {
        if err := p.extendList(ctx, v); err != nil {
            return fmt.Errorf("extend list error: %w", err)
        }
        if v.Name != "" {
            ctx.listRefs[v.Name] = p.copyListRecord(v)
        }
    }
    return nil
}

func (p *redisYamlParser) buildZSets(ctx *context, data *ZSets) error {
    if data == nil {
        return nil
    }
    for _, v := range data.Values {
        if err := p.extendZSet(ctx, v); err != nil {
            return fmt.Errorf("extend zset error: %w", err)
        }
        if v.Name != "" {
            ctx.zsetRefs[v.Name] = p.copyZSetRecord(v)
        }
    }
    return nil
}

func (p *redisYamlParser) Parse(ctx *context, filename string) (*Fixture, error) {
    data, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }

    var fixture Fixture
    if err := yaml.Unmarshal(data, &fixture); err != nil {
        return nil, err
    }

    for _, parentFixture := range fixture.Inherits {
        _, err := p.fileParser.ParseFiles(ctx, []string{parentFixture})
        if err != nil {
            return nil, err
        }
    }

    if err = p.buildTemplate(ctx, fixture); err != nil {
        return nil, err
    }

    for _, databaseData := range fixture.Databases {
        if err := p.buildKeys(ctx, databaseData.Keys); err != nil {
            return nil, err
        }
        if err := p.buildMaps(ctx, databaseData.Hashes); err != nil {
            return nil, err
        }
        if err := p.buildSets(ctx, databaseData.Sets); err != nil {
            return nil, err
        }
        if err := p.buildLists(ctx, databaseData.Lists); err != nil {
            return nil, err
        }
        if err := p.buildZSets(ctx, databaseData.ZSets); err != nil {
            return nil, err
        }
    }

    return &fixture, nil
}
