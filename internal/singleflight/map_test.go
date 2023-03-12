package singleflight

import (
	"testing"
	"time"
)

func TestMap(t *testing.T) {
	t.Run("", func(t *testing.T) {
		m := New[int, int]()

		val, err := m.Do(1, fetch(1))
		if err != nil || val != 1 {
			t.Fatal()
		}

		val, err = m.Do(1, func() (int, error) {
			t.Fatal("value should be cached")
			return 1, nil
		})
		if err != nil || val != 1 {
			t.Fatal()
		}
	})

	t.Run("", func(t *testing.T) {
		m := New[int, int]()

		for i := 0; i < 10; i++ {
			go func() {
				val, err := m.Do(1, fetch(1))
				if err != nil || val != 1 {
					t.Fail()
				}
			}()
		}
		time.Sleep(10 * time.Millisecond)
		val, err := m.Do(1, func() (int, error) {
			t.Fatal("value should be cached")
			return 1, nil
		})
		if err != nil || val != 1 {
			t.Fatal()
		}
	})
}

func TestMapClear(t *testing.T) {
	m := New[int, int]()

	for i := 0; i < 10; i++ {
		go func(i int) {
			val, err := m.Do(i, fetch(1))
			if err != nil || val != 1 {
				t.Fail()
			}
		}(i)
	}

	time.Sleep(10 * time.Millisecond)
	if len(m.m) != 10 {
		t.Fatal()
	}
	m.Clear()
	if len(m.m) != 0 {
		t.Fatal()
	}
}

func fetch(i int) func() (int, error) {
	return func() (int, error) {
		return i, nil
	}
}
