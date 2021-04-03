package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Wrapper struct {
	Experiments []Experiment
}
type Experiment struct {
	Key       string
	File      string
	Operation string
}

type PExperiment struct {
	File      string
	Operation string
	active    bool
	backup    bool
}

func main() {
	wrapper, err := readConf("experiments.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var rFlag = flag.Bool("r", false, "restore")
	flag.Parse()
	if *rFlag {
		fmt.Println("restoring")
		err = restoreFiles()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	exps := make(map[string]*PExperiment)

	for _, e := range wrapper.Experiments {
		p := &PExperiment{File: e.File, Operation: e.Operation}
		exps[e.Key] = p
	}

	fmt.Println("parsing args")

	anyActive := false

	backups := make(map[string]bool)

	for _, arg := range os.Args[1:] {
		exps[arg].active = true
		anyActive = true
		backups[exps[arg].File] = true
	}

	if !anyActive {
		log.Fatal("No experiments performed!")
	}

	fmt.Println("creating backups")

	for k, _ := range backups {
		err = backupFiles(k)
		if err != nil {
			log.Fatal(err)
		}
	}

	restorable := make([]string, 0)

	for _, exp := range exps {
		if !exp.active {
			continue
		}

		restorable = append(restorable, exp.File)

		ops := strings.Split(exp.Operation, " ")
		cmd := exec.Command(ops[0], ops[1], exp.File)
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
	}

	wr, err := yaml.Marshal(&restorable)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("./restore.yaml", wr, 0644)
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

func readConf(filename string) (*Wrapper, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &Wrapper{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %v", filename, err)
	}

	return c, nil
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