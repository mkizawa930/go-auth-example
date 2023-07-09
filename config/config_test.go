package config

import "testing"

func TestConfig(t *testing.T) {
	want := 1
	got := 1
	if want != got {
		t.Fail()
	}
}

// func TestGetProviderConfig(t *testing.T) {
// 	dirpath, err := os.Getwd()
// 	if err != nil {
// 		t.Log(dirpath)
// 		t.Fail()
// 	}
// 	filepath := dirpath + "/" + "config.yaml"
// 	t.Log(filepath)
// 	b, err := os.ReadFile(filepath)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	var m map[string]Config
// 	err = yaml.Unmarshal(b, &m)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	got, ok := m["dev"]
// 	if ok {
// 		t.Fail()
// 	}
// 	provider, ok := got.ProviderConfigMap["google"]
// 	if !ok {
// 		t.Fail()
// 	}
// 	if provider.Issuer != "https://accounts.google.com" {
// 		t.Log(provider.Issuer)
// 		t.Fail()
// 	}
// }
