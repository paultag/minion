package main

import (
	"time"
)

func YakYakYakGetAJob(wait time.Duration, channels ...<-chan Job) Job {
	for {
		for _, channel := range channels {
			el := ReadTimeout(wait, channel)
			if el != nil {
				return *el
			}
		}
	}
}

func ReadTimeout(wait time.Duration, channel <-chan Job) *Job {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(wait * time.Second)
		timeout <- true
	}()

	select {
	case job := <-channel:
		return &job
	case <-timeout:
		return nil
	}
}
