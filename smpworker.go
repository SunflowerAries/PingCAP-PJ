package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type StateType int32

const (
	WAITER StateType = 0
	WORKER StateType = 1
)

const (
	hashLimit = 101
)

type Record struct {
	a, b int
}

type Records struct {
	records []Record
	mu      sync.Mutex
}

type Count struct {
	mu       sync.Mutex
	count    int
	cntLimit int
	cntCh    chan bool
}

type Worker struct {
	records []Records
	other   *Waiter
	synCh   chan bool // between worker and waiter
	cnt     Count
}

type Waiter struct {
	other   *Worker
	synCh   chan bool // between worker and waiter
	finCh   chan int  // indicate both have finished
	cnt     Count
	matches int
}

func fileSeparate(filePrefix string, fileNames []string, regex *regexp.Regexp) ([]string, []string) {
	wanted := []string{}
	rest := []string{}
	for _, fileName := range fileNames {
		if match := regex.MatchString(fileName); match {
			wanted = append(wanted, filepath.Join(filePrefix, fileName))
		} else {
			rest = append(rest, filepath.Join(filePrefix, fileName))
		}
	}
	return wanted, rest
}

func fileReader(fileName string) [hashLimit][]Record {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	records := [hashLimit][]Record{}
	for scanner.Scan() {
		split := strings.Split(scanner.Text(), "\t")
		record := Record{}
		record.a, err = strconv.Atoi(split[0])
		if err != nil {
			log.Fatal(err)
		}
		record.b, err = strconv.Atoi(split[1])
		if err != nil {
			log.Fatal(err)
		}
		records[record.a%hashLimit] = append(records[record.a%hashLimit], record)
	}
	return records
}

func (w *Worker) startWorker(fileNames []string) {
	cntCh := make(chan bool)
	w.cnt = Count{count: 0, cntLimit: len(fileNames), cntCh: cntCh}
	for _, fileName := range fileNames {
		go func(fileName string) {
			records := fileReader(fileName)
			for i := 0; i < hashLimit; i++ {
				w.records[i].mu.Lock()
				w.records[i].records = append(w.records[i].records, records[i]...)
				w.records[i].mu.Unlock()
			}

			w.cnt.mu.Lock()
			defer w.cnt.mu.Unlock()
			if w.cnt.count++; w.cnt.count >= w.cnt.cntLimit {
				w.cnt.cntCh <- true
			}
		}(fileName)
	}
	<-w.cnt.cntCh   // block until finish all reads from file
	w.synCh <- true // to waiter
}

func (w *Waiter) startWaiter(fileNames []string) {
	cntCh := make(chan bool)
	startCh := make(chan bool)
	fileLen := len(fileNames)
	w.cnt = Count{count: 0, cntLimit: fileLen, cntCh: cntCh}

	for _, fileName := range fileNames {
		go func(fileName string) {
			records := fileReader(fileName)
			<-startCh
			matches := 0
			for idx := 0; idx < hashLimit; idx++ {
				for _, record := range records[idx] {
					for _, counterpart := range w.other.records[idx].records {
						if record.a == counterpart.a && record.b > counterpart.b {
							matches++
						}
					}
				}
			}
			w.cnt.mu.Lock()
			defer w.cnt.mu.Unlock()
			w.matches += matches
			if w.cnt.count++; w.cnt.count >= w.cnt.cntLimit {
				w.finCh <- w.matches
				w.cnt.cntCh <- true
			}
		}(fileName)
	}
	<-w.synCh
	for i := 0; i < fileLen; i++ {
		startCh <- true
	}
	<-w.cnt.cntCh
}

func start(datadir string) {
	pwd, _ := os.Getwd()
	dataFilePath := filepath.Join(pwd, datadir)
	filePathNames, err := ioutil.ReadDir(dataFilePath)
	if err != nil {
		log.Fatal(err)
	}

	fileNames := []string{}
	for i := range filePathNames {
		fileNames = append(fileNames, filePathNames[i].Name())
	}

	regex, _ := regexp.Compile("t1")
	t1Files, t2Files := fileSeparate(dataFilePath, fileNames, regex)

	synCh := make(chan bool)
	finCh := make(chan int)
	w1 := &Worker{synCh: synCh}
	w2 := &Waiter{synCh: synCh, finCh: finCh}
	w1.other = w2
	w2.other = w1

	for i := 0; i < hashLimit; i++ {
		w1.records = append(w1.records, Records{})
	}

	if len(t1Files) < len(t2Files) {
		go w1.startWorker(t1Files)
		go w2.startWaiter(t2Files)
	} else {
		go w1.startWorker(t2Files)
		go w2.startWaiter(t1Files)
	}
	fmt.Println(<-w2.finCh)
}
