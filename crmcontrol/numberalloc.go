package crmcontrol

import "errors"

const (
	MAX_LOOPS = 32
)

// Adapted for a fixed [1, 32767] range from LINSTOR com.linbit.NumberAlloc
func GetFreeTargetId(sortedTidList []int16) (int16, bool) {
	var freeTid int16
	haveFreeTid := false
	tidCount := len(sortedTidList)
	if tidCount < 32767 {
		if tidCount > 0 {
			startIdx := 0
			endIdx := tidCount
			resultIdx := -1
			loopGuard := 0
			for startIdx < endIdx {
				if loopGuard >= MAX_LOOPS {
					panic(errors.New("Implementation error: Loop guard triggered, possible use of unsorted array with GetFreeTargetId()"))
				}
				width := endIdx - startIdx
				midIdx := startIdx + (width >> 1)
				if sortedTidList[midIdx] == int16(midIdx+1) {
					// No gap in the lower part of the current region
					// Isolate higher part of the region
					startIdx = midIdx + 1
				} else {
					// Gap somewhere in the lower part of the region
					// Isolate lower part of the region
					endIdx = midIdx
					resultIdx = midIdx
				}
				loopGuard++
			}

			if resultIdx > 0 {
				freeTid = sortedTidList[resultIdx-1] + 1
				haveFreeTid = true
			} else if resultIdx == 0 {
				freeTid = 1
				haveFreeTid = true
			} else {
				// Greater numbers than the occupied ones are available
				freeTid = sortedTidList[tidCount-1] + 1
				haveFreeTid = true
			}
		} else {
			freeTid = 1
			haveFreeTid = true
		}
	}
	return freeTid, haveFreeTid
}
