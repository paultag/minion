package minion

import (
	"fmt"
	"io/ioutil"
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
