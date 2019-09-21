package main

type store interface {
	Add(data []byte)
}
