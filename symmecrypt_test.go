package symmecrypt_test

import (
	"testing"

	"github.com/ovh/configstore"
	"github.com/ovh/symmecrypt"
	"github.com/ovh/symmecrypt/keyloader"
)

func ProviderTest() (configstore.ItemList, error) {
	ret := configstore.ItemList{
		Items: []configstore.Item{
			configstore.NewItem(
				keyloader.EncryptionKeyConfigName,
				`{"key":"5fdb8af280b007a46553dfddb3f42bc10619dcabca8d4fdf5239b09445ab1a41","identifier":"test","sealed":false,"timestamp":1522325806,"cipher":"aes-gcm"}`,
				1,
			),
			configstore.NewItem(
				keyloader.EncryptionKeyConfigName,
				`{"key":"7db2b4b695e11563edca94b0f9c7ad16919fc11eac414c1b1706cbaa3c3e61a4b884301ae4e8fbedcc4f000b9c52904f13ea9456379d373524dea7fef79b39f7","identifier":"test-composite","sealed":false,"timestamp":1522325758,"cipher":"aes-pmac-siv"}`,
				1,
			),
			configstore.NewItem(
				keyloader.EncryptionKeyConfigName,
				`{"key":"95371d0966180e05a67aa132669001061b57d423aeec83c49d18d32347e3d335","identifier":"test-composite","sealed":false,"timestamp":1522325802,"cipher":"chacha20-poly1305"}`,
				1,
			),
		},
	}
	return ret, nil
}

func TestMain(m *testing.M) {

	configstore.RegisterProvider("test", ProviderTest)

	m.Run()
}

func TestEncryptDecrypt(t *testing.T) {
	text := []byte("eoeodecrytp")

	extra := []byte("aa")
	extra2 := []byte("bb")

	k, err := keyloader.LoadKey("test")
	if err != nil {
		t.Fatal(err)
	}

	encr, err := k.Encrypt(text)
	if err != nil {
		t.Fatal(err)
	}

	encrExtra, err := k.Encrypt(text, extra, extra2)
	if err != nil {
		t.Fatal(err)
	}

	decr, err := k.Decrypt(encr)
	if err != nil {
		t.Fatal(err)
	}

	decrExtra, err := k.Decrypt(encrExtra, extra, extra2)
	if err != nil {
		t.Fatal(err)
	}

	_, err = k.Decrypt(encrExtra)
	if err == nil {
		t.Fatal("successfully decrypted cipher+extra without using extra -> ERROR")
	}

	_, err = k.Decrypt(encrExtra, []byte("cc"), []byte("dd"))
	if err == nil {
		t.Fatal("successfully decrypted cipher+extra using wrong extra -> ERROR")
	}

	_, err = k.Decrypt(encr, extra, extra2)
	if err == nil {
		t.Fatal("succerssfully decrypted cipher while using extra data -> ERROR")
	}

	if string(decr) != string(text) {
		t.Errorf("not equal when decrypt text encrypted,  %s != %s", text, decr)
	}

	if string(decrExtra) != string(text) {
		t.Errorf("not equal when decrypt text encrypted [extra data],  %s != %s", text, decrExtra)
	}
}

type testObfuscate struct {
	Name   string
	Amount int
}

func TestEncryptDecryptMarshal(t *testing.T) {

	k, err := keyloader.LoadKey("test")
	if err != nil {
		t.Fatal(err)
	}

	origin := &testObfuscate{
		Name:   "test",
		Amount: 10,
	}

	extra := []byte("aa")
	extra2 := []byte("bb")

	r, err := k.EncryptMarshal(origin)
	if err != nil {
		t.Fatal(err)
	}

	rExtra, err := k.EncryptMarshal(origin, extra, extra2)
	if err != nil {
		t.Fatal(err)
	}

	target := &testObfuscate{}
	targetExtra := &testObfuscate{}

	err = k.DecryptMarshal(r, target)
	if err != nil {
		t.Fatal(err)
	}

	err = k.DecryptMarshal(rExtra, targetExtra, extra, extra2)
	if err != nil {
		t.Fatal(err)
	}

	err = k.DecryptMarshal(rExtra, targetExtra)
	if err == nil {
		t.Fatal("succerssfully decrypted cipher without using extra data -> ERROR")
	}

	if target.Name != origin.Name || target.Amount != origin.Amount {
		t.Errorf("Not same deobfuscated result %s, %d", target.Name, target.Amount)
	}
	if targetExtra.Name != origin.Name || targetExtra.Amount != origin.Amount {
		t.Errorf("Not same deobfuscated result %s, %d", targetExtra.Name, targetExtra.Amount)
	}
}

func TestCompositeKey(t *testing.T) {

	kC, err := keyloader.LoadKey("test-composite")
	if err != nil {
		t.Fatal(err)
	}

	var k, k2 symmecrypt.Key

	comp, ok := kC.(symmecrypt.CompositeKey)
	if !ok {
		t.Fatal("Expected a composite key instance")
	}

	if len(comp) < 2 {
		t.Fatalf("composite len should be 2, got %d", len(comp))
	}

	k = comp[0]
	k2 = comp[1]

	text := []byte("eoeodecrytp")

	encr, err := kC.Encrypt(text)
	if err != nil {
		t.Fatal(err)
	}

	decr, err := kC.Decrypt(encr)
	if err != nil {
		t.Fatal(err)
	}
	if string(text) != string(decr) {
		t.Errorf("not equal when decrypt text encrypted,  %s != %s", text, decr)
	}

	decr1, err := k.Decrypt(encr)
	if err != nil {
		t.Fatal(err)
	}
	if string(text) != string(decr1) {
		t.Errorf("not equal when decrypt text encrypted,  %s != %s", text, decr1)
	}

	_, err = k2.Decrypt(encr)
	if err == nil {
		t.Fatal("successfully decrypted composite encrypt result with low-priority key -> ERROR")
	}

	encr2, err := k2.Encrypt(text)
	if err != nil {
		t.Fatal(err)
	}

	decr2, err := kC.Decrypt(encr2)
	if err != nil {
		t.Fatal(err)
	}

	if string(text) != string(decr2) {
		t.Errorf("not equal when decrypt text encrypted,  %s != %s", text, decr2)
	}

	extra := []byte("aa")
	extra2 := []byte("bb")

	encr3, err := k.Encrypt(text, extra, extra2)
	if err != nil {
		t.Fatal(err)
	}

	decr3, err := kC.Decrypt(encr3, extra, extra2)
	if err != nil {
		t.Fatal(err)
	}
	if string(text) != string(decr3) {
		t.Errorf("not equal when decrypt text encrypted,  %s != %s", text, decr3)
	}

	_, err = kC.Decrypt(encr3)
	if err == nil {
		t.Fatal("successfully decrypted cipher+extra without using extra -> ERROR")
	}
}
