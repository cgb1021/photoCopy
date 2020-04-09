package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

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
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
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
		fmt.Printf("mkdir dest failed![%v]\n", err)
		return
	}
	// 创建临时目录
	err = mkdir(temp)
	if err != nil {
		fmt.Printf("mkdir temp failed![%v]\n", err)
		return
	}
	// 获取文件
	files, _ := getFiles(src)
	PthSep := string(os.PathSeparator)
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println(err)
		} else {
			bytes := md5.Sum(data)
			val := fmt.Sprintf("%x", bytes)
			tmpPath := temp + PthSep + val
			exists, _ := pathExists(tmpPath)
			if !exists {
				_, err := copy(file, dest+PthSep+val+"."+filepath.Base(file))
				if err != nil {
					fmt.Printf("err:%s", file)
				} else {
					f, _ := os.Create(tmpPath)
					f.Close()
					fmt.Printf("done %s\n", file)
				}
			}
		}
	}
	os.RemoveAll(temp)
	fmt.Println("====================================\ncomplete!")
}
