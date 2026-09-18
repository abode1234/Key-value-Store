# write in file path 

when we save a data in disk we give it the file path this maybe not save 
This naive approach has some drawbacks:
1. It truncates the file before updating it. What if the file needs to be read concurrently?
2. Writing data to files may not be atomic, depending on the size of the write. Con-
current readers might get incomplete data.
3. When is the data actually persisted to the disk? The data is probably still in the
operating system’s page cache after the write syscall returns. What’s the state of the
file when the system crashes and reboots?


`golang

func SaveData(path string, data []byte) error{
	
	fp , err := os.OpenFile(path, os.O_WRONLY| os.O_CREATE |   os.O_EXCL, 0664)
	if err != nil {
		return err
	}
	defer fp.Close()

	_,err =fp.Write(data)
	return err
}
`
## Atomic

To address some of these problems, let’s propose a better approach

`go
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
`

This approach is slightly more sophisticated, it first dumps the data to a temporary file,
then rename the temporary file to the target file. This seems to be free of the non-atomic
problem of updating a file directly — the rename operation is atomic. If the system crashed
before renaming, the original file remains intact, and applications have no problem reading
the file concurrently.
However, this is still problematic because it doesn’t control when the data is persisted to
the disk, and the metadata (the size of the file) may be persisted to the disk before the data,
potentially corrupting the file after when the system crash. (You may have noticed that
some log files have zeros in them after a power failure, that’s a sign of file corruption.

## Atomic + file sync

To fix the problem, we must flush the data to the disk before renaming it. The Linux
syscall for this is “fsync”.

`go
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
`
Are we done yet? The answer is no. We have flushed the data to the disk, but what about
the metadata? Should we also call the fsync on the directory containing the file?
This rabbit hole is quite deep and that’s why databases are preferred over files for persisting
data to the disk

## Append-Only Logs
In some use cases, it makes sense to persist data using an append-only log

In some use cases, it makes sense to persist data using an append-only log.
`
func LogCreate(path string) (*os.File, error) {
return os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0664)
}
func LogAppend(fp *os.File, line string) error {
buf := []byte(line)
buf = append(buf, '\n')
_, err := fp.Write(buf)
if err != nil {
return err
}
return fp.Sync() // fsync
}

The nice thing about the append-only log is that it does not modify the existing data, nor

does it deal with the rename operation, making it more resistant to corruption. But logs
alone are not enough to build a database.
1. A database uses additional “indexes” to query the data efficiently. There are only
brute-force ways to query a bunch of records of arbitrary order.
2. How do logs handle deleted data? They cannot grow forever
