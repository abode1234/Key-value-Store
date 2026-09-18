package main

import (
	"fmt"
	"os"
	"path"
	"time"
)

// without atomic
func SaveData(path string, data []byte) error{
	
	fp , err := os.OpenFile(path, os.O_WRONLY| os.O_CREATE | os.O_EXCL, 0664)
	if err != nil {
		return err
	}
	defer fp.Close()

	_,err =fp.Write(data)
	return err
}


func randomInt() int64 {
	return time.Now().UnixNano()
}
// with atomic
func SaveData2(path string, data []byte) error{
	// create a tmp file with random name
	tmp := fmt.Sprintf("%s.tmp.%d", path, randomInt())	
	fp , err := os.OpenFile(tmp, os.O_WRONLY| os.O_CREATE | os.O_EXCL, 0664)
	if err != nil {
		return err
	}
	defer fp.Close()

	// write data to file path 
	_,err =fp.Write(data)
	if err != nil {
		os.Remove(tmp)
		return err
	}
	// rename tmp file to path
	return os.Rename(tmp, path)
}

// with File Sync
func SaveData3(path string, data []byte) error{
	// create a tmp file with random name
	tmp := fmt.Sprintf("%s.tmp.%d", path, randomInt())	
	fp , err := os.OpenFile(tmp, os.O_WRONLY| os.O_CREATE | os.O_EXCL, 0664)
	if err != nil {
		return err
	}
	defer fp.Close()

	// write data to file path 
	_,err =fp.Write(data)
	if err != nil {
		os.Remove(tmp)
		return err
	}
	err = fp.Sync()
	if err != nil {
		os.Remove(tmp)
		return err
	}

	// rename tmp file to path

	return os.Rename(tmp, path)
}

func LogCreate(path string) (*os.File, error) {
	return  os.OpenFile(path, os.O_WRONLY| os.O_CREATE | os.O_APPEND, 0664)
}

func LogAppend(fp *os.File, line string) error {
	buf := []byte(line)
	buf = append(buf, '\n')
	_,err :=fp.Write(buf)
	if err != nil {
		return err
	}
	return fp.Sync()
}


func main() {
	dir := "data"
	if err := os.MkdirAll(dir, 0755); err != nil {
		panic(err)
	}
	path := path.Join(dir, "kvs")
	data := []byte("hello world \n")
	
	fmt.Println("without atomic")
	err := SaveData(path, data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("with atomic")
	err = SaveData2(path, data)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("with File Sync")
	err = SaveData3(path, data)
	if err != nil {
		fmt.Println(err)
	}

	fp, err := LogCreate(path)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer fp.Close()
	LogAppend(fp, "hello world \n")
}


