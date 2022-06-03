package eth2

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/NethermindEth/posmoni/internal/utils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type initTestCase struct {
	yml     string
	file    string
	env     map[string]string
	args    []CfgChecker
	want    eth2Config
	isError bool
}

const (
	skip = "@@@SKIP@@@"
)

func setupYML(dir, yml string) (string, error) {
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

func setupInitTestCase(t *testing.T, tempDir string, tc *initTestCase) {
	if tc.yml != skip {
		f, err := setupYML(tempDir, tc.yml)
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
		viper.SetEnvPrefix("PM")
		for k, v := range tc.env {
			t.Setenv(k, v)
		}
	}
}

func cleanInitTestCase() {
	viper.Reset()
}

func TestInit(t *testing.T) {
	td := t.TempDir()

	tcs := []initTestCase{
		{
			yml: `
            validators: [123123, "0x1414fa980b"]
            consensus:
              - "http://153.168.127.111:5052"
              - "http://154.168.127.221:5052"
            execution: 
              - "http://133.168.127.111:8545"
              - "http://122.168.127.221:8545"`,
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
				{Key: Execution, ErrMsg: NoExecutionFoundError},
			},
			want: eth2Config{
				validators: []string{"123123", "0x1414fa980b"},
				consensus:  []string{"http://153.168.127.111:5052", "http://154.168.127.221:5052"},
				execution:  []string{"http://133.168.127.111:8545", "http://122.168.127.221:8545"},
			},
			isError: false,
		},
		{
			yml: `
            validators: "0x1414fa980b"
            consensus: "http://153.168.127.111:5052"
            execution: "http://122.168.127.221:8545"`,
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
				{Key: Execution, ErrMsg: NoExecutionFoundError},
			},
			want: eth2Config{
				validators: []string{"0x1414fa980b"},
				consensus:  []string{"http://153.168.127.111:5052"},
				execution:  []string{"http://122.168.127.221:8545"},
			},
			isError: false,
		},
		{
			yml: `
            validators: "0x1414fa980b"
            consensus: "http://153.168.127.111:5052"
            execution: "http://122.168.127.221:8545"`,
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
			},
			want: eth2Config{
				validators: []string{"0x1414fa980b"},
				consensus:  []string{"http://153.168.127.111:5052"},
			},
			isError: false,
		},
		{
			yml: `
            validators: 
            consensus: "http://153.168.127.111:5052"
            execution: `,
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
				{Key: Execution, ErrMsg: NoExecutionFoundError},
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml: `
            validators: 
            consensus: "http://153.168.127.111:5052"
            execution: `,
			args: []CfgChecker{
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
			},
			want: eth2Config{
				consensus: []string{"http://153.168.127.111:5052"},
			},
			isError: false,
		},
		{
			yml: `
            validators: "0x1414fa980b"
            consensus: 
            execution: `,
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
				{Key: Execution, ErrMsg: NoExecutionFoundError},
			},
			want: eth2Config{
				validators: []string{"0x1414fa980b"},
			},
			isError: true,
		},
		{
			yml: `
            validators: "0x1414fa980b"
            consensus: 
            execution: `,
			args: []CfgChecker{
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Execution, ErrMsg: NoExecutionFoundError},
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml: `
            consensus: `,
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
				{Key: Execution, ErrMsg: NoExecutionFoundError},
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml: `
            consensus: `,
			args: []CfgChecker{
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml: ``,
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
				{Key: Execution, ErrMsg: NoExecutionFoundError},
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml:     ``,
			args:    []CfgChecker{},
			want:    eth2Config{},
			isError: false,
		},
	}

	for i, tc := range tcs {
		name := fmt.Sprintf("Test case with config file %d", i)
		t.Run(name, func(t *testing.T) {
			setupInitTestCase(t, td, &tc)

			viper.AddConfigPath(td)
			viper.SetConfigType("yml")
			viper.SetConfigName(tc.file)

			got, err := Init(tc.args)

			descr := fmt.Sprintf("Init() with yml %s", tc.yml)
			if err = utils.CheckErr(descr, tc.isError, err); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.want, got, descr)

			cleanInitTestCase()
		})
	}

	tcs = []initTestCase{
		{
			yml: skip,
			env: map[string]string{
				"PM_VALIDATORS": "123123,0x1414fa980b",
				"PM_CONSENSUS":  "http://153.168.127.111:5052,http://154.168.127.221:5052",
				"PM_EXECUTION":  "http://133.168.127.111:8545,http://122.168.127.221:8545",
			},
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
				{Key: Execution, ErrMsg: NoExecutionFoundError},
			},
			want: eth2Config{
				validators: []string{"123123", "0x1414fa980b"},
				consensus:  []string{"http://153.168.127.111:5052", "http://154.168.127.221:5052"},
				execution:  []string{"http://133.168.127.111:8545", "http://122.168.127.221:8545"},
			},
			isError: false,
		},
		{
			yml: skip,
			env: map[string]string{
				"PM_VALIDATORS": "0x1414fa980b",
				"PM_CONSENSUS":  "http://153.168.127.111:5052",
				"PM_EXECUTION":  "http://133.168.127.111:8545",
			},
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
				{Key: Execution, ErrMsg: NoExecutionFoundError},
			},
			want: eth2Config{
				validators: []string{"0x1414fa980b"},
				consensus:  []string{"http://153.168.127.111:5052"},
				execution:  []string{"http://133.168.127.111:8545"},
			},
			isError: false,
		},
		{
			yml: skip,
			env: map[string]string{
				"PM_VALIDATORS": "0x1414fa980b",
				"PM_CONSENSUS":  "http://153.168.127.111:5052",
				"PM_EXECUTION":  "http://133.168.127.111:8545",
			},
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Execution, ErrMsg: NoExecutionFoundError},
			},
			want: eth2Config{
				validators: []string{"0x1414fa980b"},
				execution:  []string{"http://133.168.127.111:8545"},
			},
			isError: false,
		},
		{
			yml: skip,
			env: map[string]string{
				"PM_VALIDATORS": "",
				"PM_CONSENSUS":  "http://153.168.127.111:5052",
			},
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
				{Key: Execution, ErrMsg: NoExecutionFoundError},
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml: skip,
			env: map[string]string{
				"PM_VALIDATORS": "",
				"PM_CONSENSUS":  "http://153.168.127.111:5052",
			},
			args: []CfgChecker{
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
			},
			want: eth2Config{
				consensus: []string{"http://153.168.127.111:5052"},
			},
			isError: false,
		},
		{
			yml: skip,
			env: map[string]string{
				"PM_CONSENSUS": "http://153.168.127.111:5052",
			},
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
				{Key: Execution, ErrMsg: NoExecutionFoundError},
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml: skip,
			env: map[string]string{
				"PM_CONSENSUS": "http://153.168.127.111:5052",
			},
			args: []CfgChecker{
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
			},
			want: eth2Config{
				consensus: []string{"http://153.168.127.111:5052"},
			},
			isError: false,
		},
		{
			yml: skip,
			env: map[string]string{
				"PM_VALIDATORS": "0x1414fa980b",
			},
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
				{Key: Execution, ErrMsg: NoExecutionFoundError},
			},
			want: eth2Config{
				validators: []string{"0x1414fa980b"},
			},
			isError: true,
		},
		{
			yml: skip,
			env: map[string]string{
				"PM_VALIDATORS": "0x1414fa980b",
				"PM_CONSENSUS":  "",
			},
			args: []CfgChecker{
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml: skip,
			env: map[string]string{
				"PM_VALIDATORS": "",
				"PM_CONSENSUS":  "",
			},
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml: skip,
			env: map[string]string{},
			args: []CfgChecker{
				{Key: Validators, ErrMsg: NoValidatorsFoundError},
				{Key: Consensus, ErrMsg: NoConsensusFoundError},
				{Key: Execution, ErrMsg: NoExecutionFoundError},
			},
			want:    eth2Config{},
			isError: true,
		},
		{
			yml:     skip,
			env:     map[string]string{},
			args:    []CfgChecker{},
			want:    eth2Config{},
			isError: false,
		},
	}

	for i, tc := range tcs {
		name := fmt.Sprintf("Test case with ENV %d", i)
		t.Run(name, func(t *testing.T) {
			setupInitTestCase(t, td, &tc)

			got, err := Init(tc.args)

			descr := fmt.Sprintf("Init() with env %s", tc.env)
			if err = utils.CheckErr(descr, tc.isError, err); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tc.want, got, descr)
		})
	}
}
