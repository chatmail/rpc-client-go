package deltachat

import (
	"os"
	"testing"
)

var acfactory *AcFactory

func TestMain(m *testing.M) {
	acfactory = &AcFactory{Debug: os.Getenv("TEST_DEBUG") == "1"}
	acfactory.TearUp()
	defer acfactory.TearDown()
	m.Run()
}

func strptr(s string) *string { return &s }
