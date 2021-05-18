package main

import (
	"os"
	"sync"
)

type FileRW interface {
	Read(file *File) (bytes []byte, err error)
	Write(file *File, bytes []byte) (err error)
}

type File struct {
	File *os.File
	Rw sync.RWMutex
	Cond *sync.Cond
	IsEof uint32
}

type Read struct {
	Offset uint64    //偏移量
	ReadLen uint64   //读取长度
	Rsn uint64      //读取计数

}

type Write struct {
	Offset uint64     //偏移量
	WriteLen uint64   //写入长度
	Wsn uint64        //写入计数
}





