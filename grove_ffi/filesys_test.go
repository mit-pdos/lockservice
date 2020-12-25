package grove_ffi

import "testing"
import "math/rand"
import "bytes"
import "sync"
import "fmt"

func randomByteSlice(n int64) []byte {
	s := make([]byte, n)
	rand.Read(s)
	return s
}

func tw(t *testing.T, f string, c []byte) {
	Write(f, c)
}

func tr(t *testing.T, f string, exp []byte) {
	actual := Read(f)
	if !bytes.Equal(exp, actual) {
		t.Fatalf("Read(%s) gave unexpected content; expected len = %d, actual len = %d", f, len(exp), len(actual))
	}
}

func TestReadAfterWrite(t *testing.T) {
	fmt.Print("Basic filesys test")
	num_files := 10
	num_iters := 50
	max_content_size := int64(2048)
	var wg sync.WaitGroup
	worker := func(i int) {
		defer wg.Done()
		filename := "TestFile" + fmt.Sprint(i)
		for i := 0; i < num_iters; i++ {
			content := randomByteSlice(rand.Int63n(max_content_size))
			tw(t, filename, content)
			tr(t, filename, content)
		}
	}

	for i := 0; i < num_files; i++ {
		wg.Add(1)
		go worker(i)
	}
	wg.Wait()
	fmt.Printf("  ... Passed\n")
}
