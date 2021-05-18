package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestNewBasicHoster(t *testing.T) {
	cfg := HosterConfig{port: ":8001"}
	h := NewBasicHoster(&cfg)
	h.start()
	select {
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
}

func TestSampel(t *testing.T) {
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8000", nil)
}

func Test_RWt(t *testing.T)  {
	f1, err := os.OpenFile("./text", os.O_RDWR|os.O_CREATE, 0666)
	f2, err := os.OpenFile("./text", os.O_RDWR|os.O_CREATE, 0666)
	defer f1.Close()
	defer f2.Close()
	if err != nil{
		panic(err)
	}
	r := bufio.NewReader(f1)
	w := bufio.NewWriter(f2)

	_,err = w.WriteString("hello\n")
	if err != nil{
		panic(err)
	}
	err = w.Flush()
	if err != nil{
		panic(err)
	}
	_,err = w.WriteString("Word\n")
	if err != nil{
		panic(err)
	}
	err = w.Flush()

	if err != nil{
		panic(err)
	}

	l,ok,err := r.ReadLine()
	if err != nil{
		panic(err)
	}
	fmt.Println(ok)
	fmt.Println(string(l))
	l,_,_ = r.ReadLine()
	fmt.Println(string(l))
}

func FileOpen(path string) (*os.File, error) {
	if fi, err := os.Stat(path); err == nil {
		if !fi.IsDir() {
			return nil, fmt.Errorf("open %s: not a directory", path)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0766); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	var currenttime = time.Now().Format("2006-01-02_15.04.05")

	logfile, err := os.OpenFile(path+currenttime+"_LOG.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return logfile, nil
}

func Test_Input(t *testing.T) {
	r := bufio.NewReader(os.Stdin)
	go func() {
		for true {
			l,err:= r.ReadString('\n')
			if err != nil{
				if err == io.EOF{
					continue
				}
				panic(err)
			}
			fmt.Println(l)
		}
	}()
	select {

	}
}