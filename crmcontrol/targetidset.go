package crmcontrol

import dsaext "github.com/raltnoeder/godsaext"

type TargetIdSet struct {
	tidMap *dsaext.TreeMap
}

type TargetIdIterator struct {
	tidMapIter *dsaext.TreeMapIterator
}

func NewTargetIdSet() TargetIdSet {
	return TargetIdSet{tidMap: dsaext.NewTreeMap(dsaext.CompareInt16)}
}

func (tidSet *TargetIdSet) Insert(targetId int16) bool {
	if targetId >= 1 {
		tidSet.tidMap.Insert(targetId, nil)
	}
	return targetId >= 1
}

func (tidSet *TargetIdSet) Remove(targetId int16) {
	tidSet.tidMap.Remove(targetId)
}

func (tidSet *TargetIdSet) Contains(targetId int16) bool {
	_, haveEntry := tidSet.tidMap.Get(targetId)
	return haveEntry
}

func (tidSet *TargetIdSet) GetSize() int {
	return tidSet.tidMap.GetSize()
}

func (tidSet *TargetIdSet) ToSortedArray() []int16 {
	tidArray := make([]int16, tidSet.tidMap.GetSize())
	idx := 0
	iter := tidSet.tidMap.Iterator()
	for tid, _, isValid := iter.Next(); isValid; tid, _, isValid = iter.Next() {
		tidArray[idx] = tid.(int16)
		idx++
	}
	return tidArray
}

func (tidSet *TargetIdSet) Iterator() TargetIdIterator {
	return TargetIdIterator{tidMapIter: tidSet.tidMap.Iterator()}
}

func (tidIter *TargetIdIterator) Next() (int16, bool) {
	var result int16 = 0
	key, _, haveNext := tidIter.tidMapIter.Next()
	if haveNext {
		result = key.(int16)
	}
	return result, haveNext
}
