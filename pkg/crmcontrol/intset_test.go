package crmcontrol

import (
	"math"
	"testing"
)

func BenchmarkIntSet(b *testing.B) {
	set := NewIntSet()
	for i := 1; i <= math.MaxInt16; i++ {
		set.Add(i)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		set.GetFree(1, math.MaxInt16)
	}
}

func TestFindEmpty(t *testing.T) {
	set := NewIntSet()
	if f, ok := set.GetFree(1, math.MaxInt16); !ok || f != 1 {
		t.Errorf("Expected to not find min ID in an empty set")
	}
}

func TestFindSecond(t *testing.T) {
	set := NewIntSet()
	set.Add(1)
	if f, ok := set.GetFree(1, math.MaxInt16); !ok || f != 2 {
		t.Errorf("Expected to find second element in set")
	}
}

func TestFindHole(t *testing.T) {
	set := NewIntSet()
	for i := 1; i < 25; i++ {
		if i == 23 {
			continue
		}

		set.Add(i)
	}
	if f, ok := set.GetFree(1, math.MaxInt16); !ok || f != 23 {
		t.Errorf("Expected to find hole element in set")
	}
}

func TestFindLast(t *testing.T) {
	set := NewIntSet()
	for i := 1; i <= 24; i++ {
		set.Add(i)
	}
	if f, ok := set.GetFree(1, 25); !ok || f != 25 {
		t.Errorf("Expected to find last free element in set")
	}
}

func TestFindFull(t *testing.T) {
	set := NewIntSet()
	for i := 1; i <= 25; i++ {
		set.Add(i)
	}
	if _, ok := set.GetFree(1, 25); ok {
		t.Errorf("Expected to not find ID in full set")
	}
}
