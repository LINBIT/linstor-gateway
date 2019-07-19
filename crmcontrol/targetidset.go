package crmcontrol

import dsaext "github.com/raltnoeder/godsaext"

type TargetIdSet struct {
	tidMap *dsaext.TreeMap
}

type TargetIdIterator struct {
	tidMapIter *dsaext.TreeMapIterator
}

func NewTargetIdSet() TargetIdSet {
	return TargetIdSet{tidMap: dsaext.NewTreeMap(dsaext.CompareUInt8)}
}

func (tidSet *TargetIdSet) Insert(targetId uint8) {
	tidSet.tidMap.Insert(targetId, nil)
}

func (tidSet *TargetIdSet) Remove(targetId uint8) {
	tidSet.tidMap.Remove(targetId)
}

func (tidSet *TargetIdSet) Contains(targetId uint8) bool {
	_, haveEntry := tidSet.tidMap.Get(targetId)
	return haveEntry
}

func (tidSet *TargetIdSet) GetSize() int {
	return tidSet.tidMap.GetSize()
}

func (tidSet *TargetIdSet) Iterator() TargetIdIterator {
	return TargetIdIterator{tidMapIter: tidSet.tidMap.Iterator()}
}

func (tidIter *TargetIdIterator) Next() (uint8, bool) {
	var result uint8 = 0
	key, _, haveNext := tidIter.tidMapIter.Next()
	if haveNext {
		result = key.(uint8)
	}
	return result, haveNext
}
