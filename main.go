package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

var debugLog *log.Logger

//获取指定目录下的所有文件,包含子目录下的文件
func getFiles(dirPth string) (files []string, err error) {
	var dirs []string
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}

	PthSep := string(os.PathSeparator)

	for _, fi := range dir {
		fName := fi.Name()
		if fi.IsDir() { // 目录, 递归遍历
			dirs = append(dirs, dirPth+PthSep+fName)
			getFiles(dirPth + PthSep + fName)
		} else {
			// 过滤指定格式
			reg := regexp.MustCompile(`(?i)\.(?:jpe?g|png)$`)
			ok := reg.MatchString(fName)
			if ok {
				files = append(files, dirPth+PthSep+fName)
			}
		}
	}

	// 读取子目录下文件
	for _, table := range dirs {
		temp, _ := getFiles(table)
		for _, temp1 := range temp {
			files = append(files, temp1)
		}
	}

	return files, nil
}
func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		debugLog.Printf("%s is not a regular file", src)
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		debugLog.Printf("can't open src file %s", src)
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		debugLog.Printf("can't open dest file %s", dst)
		return 0, err
	}

	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func mkdir(path string) error {
	exists, err := pathExists(path)
	if !exists {
		err = os.Mkdir(path, os.ModePerm)
		_, err = pathExists(path)
	}

	return err
}

func init() {
	logFile, err := os.Create("./debug.log")
	// defer file.Close()
	if err != nil {
		log.Fatalln("open log file error !")
		debugLog = nil
	} else {
		debugLog = log.New(logFile, "[Debug]", log.LstdFlags)
	}
}
func main() {
	src := "./src"
	dest := "./dest"
	temp := "./temp"
	if len(os.Args) > 1 {
		src = os.Args[1]
	}
	// 创建储存目录
	err := mkdir(dest)
	if err != nil {
		debugLog.Printf("mkdir dest failed![%v]", err)
		fmt.Printf("mkdir dest failed![%v]\n", err)
		return
	}
	// 创建临时目录
	err = mkdir(temp)
	if err != nil {
		debugLog.Printf("mkdir temp failed![%v]", err)
		fmt.Printf("mkdir temp failed![%v]\n", err)
		return
	}
	// 获取文件
	files, _ := getFiles(src)
	total := len(files)
	counter := 0
	duplicate := 0
	PthSep := string(os.PathSeparator)
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			debugLog.Printf("Sum md5 error![%v]", err)
			fmt.Printf("Sum md5 error![%v]", err)
		} else {
			val := fmt.Sprintf("%x", md5.Sum(data))
			tmpPath := temp + PthSep + val
			exists, _ := pathExists(tmpPath)
			if !exists {
				savePath := dest + PthSep + filepath.Base(file)
				exists, _ := pathExists(savePath)
				if exists {
					savePath = dest + PthSep + val + "." + filepath.Base(file)
				}
				_, err := copy(file, savePath)
				if err != nil {
					debugLog.Printf("copy file err %s", file)
					fmt.Printf("copy file err %s", file)
				} else {
					f, _ := os.Create(tmpPath)
					f.Close()
					counter++
				}
			} else {
				duplicate++
			}
		}
	}
	os.RemoveAll(temp)
	debugLog.Printf("DONE: total %d, copy %d, duplicate %d, fail %d\n", total, counter, duplicate, total-counter-duplicate)
	fmt.Printf("DONE: total %d, copy %d, duplicate %d, fail %d\n", total, counter, duplicate, total-counter-duplicate)
}
