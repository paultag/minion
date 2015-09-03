package minion

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func Tempdir() (func(), string, error) {
	popdir, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	name, err := ioutil.TempDir("", "minion.")
	if err != nil {
		return nil, "", err
	}
	err = os.Chdir(name)
	if err != nil {
		return nil, name, err
	}

	return func() {
		err := os.Chdir(popdir)
		if err != nil {
			fmt.Printf("Error during tmpdir cleanup!: %s", err)
		}
		err = os.RemoveAll(name)
		if err != nil {
			fmt.Printf("Error during tmpdir cleanup!: %s", err)
		}
	}, name, nil
}

func Download(url string) (func(), string, error) {
	fh, err := ioutil.TempFile("", "minion.")
	if err != nil {
		return nil, "", err
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fh.Name(), err
	}

	if resp.StatusCode != 200 {
		fh.Close()
		os.Remove(fh.Name())
		return nil, fh.Name(), fmt.Errorf(
			"Non-200 error code (%d) for %s",
			resp.StatusCode,
			url,
		)
	}

	io.Copy(fh, resp.Body)
	fh.Close()

	return func() {
		err = os.Remove(fh.Name())
		if err != nil {
			fmt.Printf("Error during file cleanup!: %s", err)
		}
	}, fh.Name(), nil
}
