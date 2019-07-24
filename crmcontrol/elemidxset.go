package crmcontrol

import dsaext "github.com/raltnoeder/godsaext"

type ElemIdxSet struct {
	idxMap *dsaext.TreeMap
}

type ElemIdxIterator struct {
	idxMapIter *dsaext.TreeMapIterator
}

func NewElemIdxSet() ElemIdxSet {
	return ElemIdxSet{idxMap: dsaext.NewTreeMap(ReverseCompareInt)}
}

func (idxSet *ElemIdxSet) Insert(idx int) {
	idxSet.idxMap.Insert(idx, nil)
}

func (idxSet *ElemIdxSet) Remove(idx int) {
	idxSet.idxMap.Remove(idx)
}

func (idxSet *ElemIdxSet) Contains(idx int) bool {
	_, haveEntry := idxSet.idxMap.Get(idx)
	return haveEntry
}

func (idxSet *ElemIdxSet) GetSize() int {
	return idxSet.idxMap.GetSize()
}

func (idxSet *ElemIdxSet) Iterator() ElemIdxIterator {
	return ElemIdxIterator{idxMapIter: idxSet.idxMap.Iterator()}
}

func (idxIter *ElemIdxIterator) Next() (int, bool) {
	var result int = 0
	key, _, haveNext := idxIter.idxMapIter.Next()
	if haveNext {
		result = key.(int)
	}
	return result, haveNext
}

func ReverseCompareInt(value1st, value2nd interface{}) int {
	num1st := value1st.(int)
	num2nd := value2nd.(int)
	var result int = 0
	if num1st < num2nd {
		result = 1
	} else if num1st > num2nd {
		result = -1
	}
	return result
}
