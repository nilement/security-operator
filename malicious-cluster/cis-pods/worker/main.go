package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

func main() {
	fmt.Println("backuping")
	err := backupFiles("/config/10-kubeadm.conf")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("changing permissions")
	cmd := exec.Command("chmod", "645", "/config/10-kubeadm.conf")
	err = cmd.Run()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("sleeping")
	time.Sleep(time.Hour)
}

func backupFiles(file string) error {
	return copyNative(file, "./backups/")
}

func copyNative(src, dst string) error {
	cmd := exec.Command("cp", "-rp", src, dst)
	return cmd.Run()
}

// func copy(src, dst string) (int64, error) {
// 	sourceFileStat, err := os.Stat(src)
// 	if err != nil {
// 		return 0, err
// 	}

// 	if !sourceFileStat.Mode().IsRegular() {
// 		return 0, fmt.Errorf("%s is not a regular file", src)
// 	}

// 	source, err := os.Open(src)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer source.Close()

// 	destination, err := os.Create(dst)
// 	if err != nil {
// 		return 0, err
// 	}
// 	defer destination.Close()
// 	nBytes, err := io.Copy(destination, source)
// 	return nBytes, err
// }
