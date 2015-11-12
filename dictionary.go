package browscap_go

type dictionary struct {
	expressions map[string][]*expression
	mapped      map[string]section

	expressionList    []*expression
	expressionLengths []float64
	ngramIndex        map[string]hitPairList

	completeData map[string]map[string]string
}

type section map[string]string

func newDictionary() *dictionary {
	return &dictionary{
		expressions: make(map[string][]*expression),
		mapped:      make(map[string]section),

		expressionList:    make([]*expression, 0, 0),
		expressionLengths: make([]float64, 0, 0),
		ngramIndex:        make(map[string]hitPairList),

		completeData: make(map[string]map[string]string),
	}
}

func (self *dictionary) buildCompleteData() {
	for name, _ := range self.mapped {
		self.completeData[name] = self.buildData(name)
	}
}

func (self *dictionary) getData(name string) map[string]string {
	return self.completeData[name]
}

func (self *dictionary) buildData(name string) map[string]string {
	res := make(map[string]string)

	if item, found := self.mapped[name]; found {
		// Parent's data
		if parentName, hasParent := item["Parent"]; hasParent {
			parentData := self.buildData(parentName)
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
