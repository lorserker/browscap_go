package browscap_go

type dictionary struct {
	expressions map[string][]*expression
	mapped      map[string]section

	expressionList    []*expression
	expressionLengths []float64
	ngramIndex        map[string]hitPairList
}

type section map[string]string

func newDictionary() *dictionary {
	return &dictionary{
		expressions: make(map[string][]*expression),
		mapped:      make(map[string]section),

		expressionList:    make([]*expression, 0, 0),
		expressionLengths: make([]float64, 0, 0),
		ngramIndex:        make(map[string]hitPairList),
	}
}

func (self *dictionary) findData(name string) map[string]string {
	res := make(map[string]string)

	if item, found := self.mapped[name]; found {
		// Parent's data
		if parentName, hasParent := item["Parent"]; hasParent {
			parentData := self.findData(parentName)
			if len(parentData) > 0 {
				for k, v := range parentData {
					if k == "Parent" {
						continue
					}
					res[k] = v
				}
			}
		}
		// It's item data
		if len(item) > 0 {
			for k, v := range item {
				if k == "Parent" {
					continue
				}
				res[k] = v
			}
		}
	}

	return res
}
