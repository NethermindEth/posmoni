package eth2

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type initTestCase struct {
	yml     string
	file    string
	env     map[string]string
	want    eth2Config
	isError bool
}

const (
	skip = "@@@SKIP@@@"
)

func prepareYML(dir, yml string) (string, error) {
	f, err := ioutil.TempFile(dir, "yml_test.*.yaml")
	if err != nil {
		return "", fmt.Errorf("Something wrong went while creating temporal yml file. Error: %s", err)
	}

	_, err = f.WriteString(yml)
	if err != nil {
		return "", fmt.Errorf("Something wrong went while writing temporal yml file. Error: %s", err)
	}

	return f.Name(), nil
}

func prepareInitTestCase(t *testing.T, tempDir string, tc *initTestCase) {
	if tc.yml != skip {
		f, err := prepareYML(tempDir, tc.yml)
		if err != nil {
			t.Fatal(err)
		}

		// Check if file was created correctly
		content, err := ioutil.ReadFile(f)
		if err != nil {
			t.Logf("using yml: %s", tc.yml)
			t.Fatalf("got an error creating tmp yml in test setup. Error: %s", err)
		}
		if string(content) != tc.yml {
			t.Logf("using yml: %s", tc.yml)
			t.Fatalf("Created file does not have desired content. Content: %s", string(content))
		}
		tc.file = f

		viper.SetConfigFile(f)
		if err := viper.ReadInConfig(); err != nil {
			t.Logf("using yml: %s", tc.yml)
			t.Fatalf("got an error reading config file. Error: %s", err)
		}
	} else {
		viper.SetEnvPrefix("PGM")
		for k, v := range tc.env {
			t.Setenv(k, v)
		}
	}
}

func cleanInitTestCase() {
	viper.Reset()
}

func checkErr(t *testing.T, descr string, isErr bool, err error) bool {
	l := err == nil && isErr
	r := err != nil && !isErr

	if l || r {
		t.Errorf("%s failed: %v", descr, err)
		return false
	}
	return true
}

func TestInit(t *testing.T) {
	td := t.TempDir()

	tcs := []initTestCase{
		{
			yml: `
            validators: [123123, "0x1414fa980b"]
            consensus:
              - "http://153.168.127.111:5052"
              - "http://154.168.127.221:5052"`,
			want: eth2Config{
				Validators: []string{"123123", "0x1414fa980b"},
				Consensus:  []string{"http://153.168.127.111:5052", "http://154.168.127.221:5052"},
			},
			isError: false,
		},
		{
			yml: `
            validators: "0x1414fa980b"
            consensus: "http://153.168.127.111:5052"`,
			want: eth2Config{
				Validators: []string{"0x1414fa980b"},
				Consensus:  []string{"http://153.168.127.111:5052"},
			},
			isError: false,
		},
		{
			yml: `
            validators: 
            consensus: "http://153.168.127.111:5052"`,
			want:    eth2Config{},
			isError: true,
		},
		{
			yml: `
            validators: "0x1414fa980b"
            consensus: `,
			want:    eth2Config{},
			isError: true,
		},
		{
			yml: `
            consensus: `,
			want:    eth2Config{},
			isError: true,
		},
		{
			yml:     ``,
			want:    eth2Config{},
			isError: true,
		},
	}

	for i, tc := range tcs {
		name := fmt.Sprintf("Test case with config file %d", i)
		t.Run(name, func(t *testing.T) {
			prepareInitTestCase(t, td, &tc)

			viper.AddConfigPath(td)
			viper.SetConfigType("yml")
			viper.SetConfigName(tc.file)

			got, err := Init()

			descr := fmt.Sprintf("Init() with yml %s", tc.yml)
			if ok := checkErr(t, descr, tc.isError, err); !ok {
				t.FailNow()
			}
			assert.Equal(t, tc.want, got, descr)

			cleanInitTestCase()
		})
	}

	tcs = []initTestCase{
		{
			yml: skip,
			env: map[string]string{
				"PGM_VALIDATORS": "123123,0x1414fa980b",
				"PGM_CONSENSUS":  "http://153.168.127.111:5052,http://154.168.127.221:5052",
			},
			want: eth2Config{
				Validators: []string{"123123", "0x1414fa980b"},
				Consensus:  []string{"http://153.168.127.111:5052", "http://154.168.127.221:5052"},
			},
			isError: false,
		},
		{
			yml: skip,
			env: map[string]string{
				"PGM_VALIDATORS": "0x1414fa980b",
				"PGM_CONSENSUS":  "http://153.168.127.111:5052",
			},
			want: eth2Config{
				Validators: []string{"0x1414fa980b"},
				Consensus:  []string{"http://153.168.127.111:5052"},
			},
			isError: false,
		},
		{
			yml: skip,
			env: map[string]string{
				"PGM_VALIDATORS": "",
				"PGM_CONSENSUS":  "http://153.168.127.111:5052",
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml: skip,
			env: map[string]string{
				"PGM_CONSENSUS": "http://153.168.127.111:5052",
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml: skip,
			env: map[string]string{
				"PGM_VALIDATORS": "0x1414fa980b",
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml: skip,
			env: map[string]string{
				"PGM_VALIDATORS": "0x1414fa980b",
				"PGM_CONSENSUS":  "",
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml: skip,
			env: map[string]string{
				"PGM_VALIDATORS": "",
				"PGM_CONSENSUS":  "",
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml:     skip,
			env:     map[string]string{},
			want:    eth2Config{},
			isError: true,
		},
	}

	for i, tc := range tcs {
		name := fmt.Sprintf("Test case with ENV %d", i)
		t.Run(name, func(t *testing.T) {
			prepareInitTestCase(t, td, &tc)

			got, err := Init()

			descr := fmt.Sprintf("Init() with env %s", tc.env)
			if ok := checkErr(t, descr, tc.isError, err); !ok {
				t.FailNow()
			}
			assert.Equal(t, tc.want, got, descr)
		})
	}
}
