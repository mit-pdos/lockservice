package grove_ffi

import "os"
import "path/filepath"
import "io/ioutil"
import "log"
import "fmt"

// filesystem+network library
const DataDir = "durable"

func panic_if_err(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// crash-atomically writes content to the file with name filename
func Write(filename string, content []byte) {
	_ = os.Mkdir("tmp", 0755)
	file, err := ioutil.TempFile("tmp", filename+"_*")
	panic_if_err(err)
	defer os.Remove(file.Name())

	for i := 0; i < len(content); {
		bytesWritten, err := file.Write(content[i:])
		panic_if_err(err)
		i += bytesWritten
	}

	_ = os.Mkdir(DataDir, 0755)
	panic_if_err(os.Rename(file.Name(), filepath.Join(DataDir, filename)))
}

// reads the contents of the file filename
func Read(filename string) []byte {
	content, err := ioutil.ReadFile(filepath.Join(DataDir, filename))
	panic_if_err(err)
	return content
}

// injective function u64 -> str
func U64ToString(i uint64) string {
	return fmt.Sprint(i)
}
