package utils

import (
	"bufio"
	"fmt"
	"os"
)

func AppendToFile(filePath, content string) error {
	return write(filePath, content, true)
}

func WriteToFile(filePath, content string) error {
	return write(filePath, content, false)
}

func DeleteFile(file string) error {
	return os.Remove(file)
}

func write(filePath, content string, isAppend bool) error {
	flag := os.O_WRONLY | os.O_CREATE
	if isAppend {
		flag = os.O_WRONLY | os.O_APPEND
	}

	file, err := os.OpenFile(filePath, flag, 0666)
	if err != nil {
		return fmt.Errorf("Fail to open the file: %s", filePath)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(content)
	if err != nil {
		return err
	}

	return writer.Flush()
}
