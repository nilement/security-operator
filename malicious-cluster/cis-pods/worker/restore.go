package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	fmt.Println("restoring")
	err := restoreFiles()

	if err != nil {
		log.Fatal(err)
	}
}

func restoreFiles() error {
	err := copyNative("./backups/10-kubeadm.conf", "/config/10-kubeadm.conf")
	return err
}

func copyNative(src, dst string) error {
	cmd := exec.Command("cp", "-rp", src, dst)
	return cmd.Run()
}
