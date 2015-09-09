package minion

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"pault.ag/go/debian/control"
	"pault.ag/go/debian/version"
)

func Filename(source string, v version.Version, arch, flavor string) string {
	// file paths don't have the epoch in them.
	version := ""
	if v.IsNative() {
		version = v.Version
	} else {
		version = fmt.Sprintf("%s-%s", v.Version, v.Revision)
	}
	return fmt.Sprintf(
		"%s_%s_%s.%s",
		source, version, arch, flavor,
	)

}

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
	channels *BuildChannelMap,
	suites []BuildableSuite,
) []chan Build {
	ret := []chan Build{}
	for _, suite := range suites {
		ret = append(ret, channels.Get(suite.GetKey()))
	}
	return ret
}

func HashFile(path string, algo hash.Hash) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(algo, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(algo.Sum(nil)), nil
}

func FakeChanges(
	date,
	source,
	binary,
	arch,
	version,
	distribution,
	urgency string,
	files []string,
) (string, error) {

	sha1FileHashes := ""
	sha256FileHashes := ""
	md5FileHashes := ""

	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			return "", err
		}
		size := stat.Size()

		md5Hash, err := HashFile(file, md5.New())
		if err != nil {
			return "", err
		}

		md5FileHashes += fmt.Sprintf(
			"\n %s %d fake extra %s",
			md5Hash,
			size,
			file,
		)

		sha1Hash, err := HashFile(file, sha1.New())
		if err != nil {
			return "", err
		}

		sha1FileHashes += fmt.Sprintf(
			"\n %s %d %s",
			sha1Hash,
			size,
			file,
		)

		sha256Hash, err := HashFile(file, sha256.New())
		if err != nil {
			return "", err
		}

		sha256FileHashes += fmt.Sprintf(
			"\n %s %d %s",
			sha256Hash,
			size,
			file,
		)
	}

	return fmt.Sprintf(`Format: 1.8
Date: %s
Source: %s
Binary: %s
Architecture: %s
Version: %s
Distribution: %s
Urgency: %s
Maintainer: Fake Maintainer <fake-maintainer@example.com>
Changed-By: Fake Maintainer <fake-maintainer@example.com>
Description:
 fake changes file to trick things into thinking something
 actually happened, even though it didn't.
Changes:
 fake changes file to trick things into thinking something
 actually happened, even though it didn't.
Checksums-Sha1: %s
Checcksums-Sha256: %s
Files: %s
`,
		date,
		source,
		binary,
		arch,
		version,
		distribution,
		urgency,
		// checksums

		sha1FileHashes,
		sha256FileHashes,
		md5FileHashes,
	), nil
}

// Take a DSC, and a log file, and create a fake .changes to upload
// the log to the archive. This relies on a reprepro extension.
func LogChangesFromDsc(logPath string, dsc control.DSC, suite, arch string) (string, error) {
	return FakeChanges(
		"Fri, 31 Jul 2015 12:53:50 -0400",
		dsc.Source,
		strings.Join(dsc.Binaries, " "),
		arch,
		dsc.Version.String(),
		suite,
		"low",
		[]string{logPath},
	)
}

func LogChangesFromChanges(logPath, changesPath, arch string) (string, error) {
	changes, err := control.ParseChangesFile(changesPath)
	if err != nil {
		return "", nil
	}

	return FakeChanges(
		"Fri, 31 Jul 2015 12:53:50 -0400",
		changes.Source,
		strings.Join(changes.Binaries, " "),
		arch,
		changes.Version.String(),
		changes.Distribution,
		changes.Urgency,
		[]string{logPath},
	)
}

func AppendLogToChanges(logPath, changesPath, arch string) error {
	changes, err := LogChangesFromChanges(logPath, changesPath, arch)
	if err != nil {
		return err
	}
	f, err := ioutil.TempFile("", "nmr.")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	_, err = f.Write([]byte(changes))
	if err != nil {
		return err
	}

	changes, err = MergeChanges(changesPath, f.Name())
	if err != nil {
		return err
	}

	fd, err := os.Create(changesPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	_, err = fd.Write([]byte(changes))
	if err != nil {
		return err
	}

	return nil
}

func MergeChanges(changes ...string) (string, error) {
	bytes, err := exec.Command("mergechanges", changes...).Output()
	return string(bytes), err
}
