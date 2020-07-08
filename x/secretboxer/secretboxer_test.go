package secretboxer

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

var passphrase = "123456"
var secretBoxer SecretBoxer

func initSecretBoxer(t *testing.T) {
	var err error
	passphrase = strconv.FormatInt(time.Now().Unix(), 10)
	fmt.Println(passphrase)
	secretBoxer, err = SecretBoxerForType("passphrase", passphrase, EncodingTypeURL)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPassphraseBoxer_WrapType(t *testing.T) {
	initSecretBoxer(t)
	if secretBoxer.WrapType() != "passphrase" {
		t.Errorf("secret boxer wrap type is failed, expecting passphrase, but actually getting %s", secretBoxer.WrapType())
	}
}

func TestPassphraseBoxer_Seal(t *testing.T) {
	initSecretBoxer(t)
	seal, err := secretBoxer.Seal([]byte("/chat/v1/channels"))
	if err != nil {
		t.Error(err)
	}

	t.Log(seal)
}

func TestPassphraseBoxer_Open(t *testing.T) {
	initSecretBoxer(t)
	//seal, err := secretBoxer.Seal([]byte("/api/v1/channels"))
	//if err != nil {
	//	t.Error(err)
	//}
	//
	//t.Log(seal)

	openSeal, err := secretBoxer.Open("OAEG7PZgaSfYXv/uySqp0NM+OJ90siyN4GS2nTESYbWSEjAoEaNXA8zwpmG/FzAbhb47uCBt3PbmIn5NQUJuFMeaopQ2Pf0s")
	if err != nil {
		t.Error(err)
	}

	if string(openSeal) != "/chat/v1/channels" {
		t.Errorf("secret boxer open seal is failed, expecting 666666, but actually getting %s", string(openSeal))
	}
}
