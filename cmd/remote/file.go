package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

type File struct {
	Name    string
	Content []byte
}

func ReadLocalFilesInDir(dirPath string) []File {
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatalf("Read Dir error: %v", err)
		return nil
	}
	entryNum := len(dirEntries)
	// 创建用于存储文件内容的通道
	fileChan := make(chan File, entryNum)
	var wg sync.WaitGroup
	wg.Add(entryNum)
	for i := 0; i < entryNum; i++ {
		if dirEntries[i].IsDir() {
			wg.Done()
		} else {
			fileName := dirEntries[i].Name()
			go func(dirPath string, fileName string) {
				defer wg.Done()
				// 读取文件内容
				filePath := JoinPath(dirPath, fileName)
				content, err := os.ReadFile(filePath)
				if err != nil {
					log.Printf("Failed to read file %s: %s\n", filePath, err)
					return
				}
				// 将文件内容发送到通道
				fileChan <- File{
					Name:    fileName,
					Content: content,
				}
			}(dirPath, fileName)
		}
	}
	wg.Wait()
	// 关闭通道，表示所有文件已经读取完毕
	close(fileChan)
	// 将文件内容存储在二维字节数组中
	var files []File
	for file := range fileChan {
		files = append(files, file)
	}
	return files
}
