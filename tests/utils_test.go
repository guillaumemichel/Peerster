package tests

import (
	"testing"

	"github.com/guillaumemichel/Peerster/utils"
)

func checkErr(err error, t *testing.T) {
	if err != nil {
		t.Error(err)
	}
}

func checkNoErr(err error, t *testing.T) {
	if err == nil {
		t.Errorf("error shoud have been thrown\n")
	}
}

func TestCheckFilename(t *testing.T) {
	err := utils.CheckFilename("correct_file.txt")
	checkErr(err, t)
	err = utils.CheckFilename("another(corret&one\".thingy")
	checkErr(err, t)
	err = utils.CheckFilename("incorrect/filename.jjj")
	checkNoErr(err, t)
}
