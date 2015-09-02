package minion

import (
	"time"
)

func nextBuildChan(channel chan Build, t time.Duration) *Build {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(t * time.Second)
		timeout <- true
	}()
	select {
	case job := <-channel:
		return &job
	case <-timeout:
		return nil
	}
}

func NextBuild(channels []chan Build, heartbeat time.Duration) Build {
	for {
		for _, channel := range channels {
			job := nextBuildChan(channel, heartbeat)
			if job != nil {
				return *job
			}
		}
	}
}

func GetBuildChannels(
	channels BuildChannelMap,
	suites []BuildableSuite,
) []chan Build {
	ret := []chan Build{}
	for _, suite := range suites {
		ret = append(ret, channels.Get(suite.GetKey()))
	}
	return ret
}
