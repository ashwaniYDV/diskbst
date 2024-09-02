package diskbst

import (
	"bytes"
	"errors"
	"path/filepath"
	"strconv"
	"testing"
)

func TestBST(t *testing.T) {
	t.Run("write/read", func(t *testing.T) {
		path := t.TempDir()
		pathName := filepath.Join(path, "test.bst")

		w, err := OpenWriter(pathName)
		if err != nil {
			t.Error(err)
			return
		}

		err = w.Put([]byte("hello"), []byte("world"))
		if err != nil {
			t.Error(err)
			return
		}

		err = w.Put([]byte("helloo"), []byte("worldd"))
		if err != nil {
			t.Error(err)
			return
		}

		err = w.Put([]byte("hell"), []byte("worl"))
		if err != nil {
			t.Error(err)
			return
		}

		err = w.Put([]byte("hellooo"), []byte("worlddd"))
		if err != nil {
			t.Error(err)
			return
		}

		w.Close()

		r, err := OpenReader(pathName)
		if err != nil {
			t.Error(err)
			return
		}

		val, err := r.Get([]byte("hello"))
		if err != nil {
			t.Error(err)
			return
		}

		if bytes.Compare(val, []byte("world")) != 0 {
			t.Fail()
			return
		}

		val, err = r.Get([]byte("hellooo"))
		if err != nil {
			t.Error(err)
			return
		}

		if bytes.Compare(val, []byte("worlddd")) != 0 {
			t.Fail()
			return
		}

		val, err = r.Get([]byte("hell"))
		if err != nil {
			t.Error(err)
			return
		}

		if bytes.Compare(val, []byte("worl")) != 0 {
			t.Fail()
			return
		}

		val, err = r.Get([]byte("blah"))
		if !errors.Is(err, errKeyNotFound) {
			t.Fail()
			return
		}

	})
}

// BenchmarkPut benchmarks the Put method
func BenchmarkPut(b *testing.B) {
	path := b.TempDir()
	pathName := filepath.Join(path, "benchmark.bst")

	w, err := OpenWriter(pathName)
	if err != nil {
		b.Fatal(err)
	}

	defer w.Close()

	for i := 0; i < b.N; i++ {
		key := []byte("key" + strconv.Itoa(i))
		value := []byte("value" + strconv.Itoa(i))

		if err := w.Put(key, value); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGet benchmarks the Get method
func BenchmarkGet(b *testing.B) {
	path := b.TempDir()
	pathName := filepath.Join(path, "benchmark.bst")

	w, err := OpenWriter(pathName)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		key := []byte("key" + strconv.Itoa(i))
		value := []byte("value" + strconv.Itoa(i))

		if err := w.Put(key, value); err != nil {
			b.Fatal(err)
		}
	}

	w.Close()

	r, err := OpenReader(pathName)
	if err != nil {
		b.Fatal(err)
	}

	defer r.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := []byte("key" + strconv.Itoa(i%1000))

		if _, err := r.Get(key); err != nil {
			b.Fatal(err)
		}
	}
}
