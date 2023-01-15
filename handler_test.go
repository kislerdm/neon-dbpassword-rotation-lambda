package main

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func Test_extractSecretObject(t *testing.T) {
	type args struct {
		v      *secretsmanager.GetSecretValueOutput
		secret any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				v: &secretsmanager.GetSecretValueOutput{
					SecretString: aws.String(`{"password":"` + placeholderPassword + `"}`),
				},
				secret: &SecretUser{},
			},
			wantErr: false,
		},
		{
			name: "unhappy path",
			args: args{
				v: &secretsmanager.GetSecretValueOutput{
					SecretString: aws.String(`{`),
				},
				secret: &SecretUser{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := extractSecretObject(tt.args.v, tt.args.secret); (err != nil) != tt.wantErr {
					t.Errorf("extractSecretObject() error = %v, wantErr %v", err, tt.wantErr)
				}
				if !tt.wantErr && tt.args.secret.(*SecretUser).Password != placeholderPassword {
					t.Errorf("extractSecretObject() failed to deserialize password")
				}
			},
		)
	}
}

type mockSecretsmanagerClient struct {
	secretAWSCurrent string

	secretByID map[string]map[string]string
}

func getSecret(m *mockSecretsmanagerClient, stage, version string) SecretUser {
	stages, ok := m.secretByID[version]
	if !ok {
		panic("no version " + version + " found")
	}

	s, ok := stages[stage]
	if !ok {
		panic("no stage " + stage + " for the version " + version + " found")
	}

	var secret SecretUser
	if err := json.Unmarshal([]byte(s), &secret); err != nil {
		panic(err)
	}

	return secret
}

func (m *mockSecretsmanagerClient) GetSecretValue(
	ctx context.Context, input *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options),
) (*secretsmanager.GetSecretValueOutput, error) {
	o := &secretsmanager.GetSecretValueOutput{
		ARN:           input.SecretId,
		VersionStages: []string{"AWSCURRENT"},
		SecretString:  &m.secretAWSCurrent,
	}

	if input.VersionId == nil {
		return o, nil
	}

	stages, ok := m.secretByID[*input.VersionId]
	if !ok {
		return nil, errors.New("no version " + *input.VersionId + " found")
	}

	stage := *input.VersionStage
	if stage == "" {
		stage = "AWSCURRENT"
	}

	s, ok := stages[stage]
	if !ok {
		return nil, errors.New(
			"no stage " + stage + " for the version " + *input.VersionId + " found",
		)
	}

	stagesK := make([]string, len(stages))
	var i uint8
	for k := range stages {
		stagesK[i] = k
		i++
	}

	o.VersionStages = stagesK
	o.SecretString = &s

	return o, nil
}

func (m *mockSecretsmanagerClient) PutSecretValue(
	ctx context.Context, input *secretsmanager.PutSecretValueInput, optFns ...func(*secretsmanager.Options),
) (*secretsmanager.PutSecretValueOutput, error) {
	versionID := *input.ClientRequestToken
	stage := input.VersionStages[0]

	if m.secretByID == nil {
		m.secretByID = map[string]map[string]string{}
	}

	if _, ok := m.secretByID[versionID]; !ok {
		m.secretByID[versionID] = map[string]string{}
	}

	m.secretByID[versionID][stage] = *input.SecretString

	return nil, nil
}

func (m *mockSecretsmanagerClient) DescribeSecret(
	ctx context.Context, input *secretsmanager.DescribeSecretInput, optFns ...func(*secretsmanager.Options),
) (*secretsmanager.DescribeSecretOutput, error) {
	//TODO implement me
	panic("implement me")
}

func (m *mockSecretsmanagerClient) UpdateSecretVersionStage(
	ctx context.Context, input *secretsmanager.UpdateSecretVersionStageInput,
	optFns ...func(*secretsmanager.Options),
) (*secretsmanager.UpdateSecretVersionStageOutput, error) {
	//TODO implement me
	panic("implement me")
}

var (
	placeholderSecretUserStr = `{
"dbname": "foo",
"user": "bar",
"host": "dev",
"project_id": "baz",
"branch_id": "br-foo",
"password": "` + placeholderPassword + `"}`

	placeholderSecretUser = SecretUser{
		User:         "bar",
		Password:     placeholderPassword,
		Host:         "dev",
		ProjectID:    "baz",
		BranchID:     "br-foo",
		DatabaseName: "foo",
	}
)

func Test_createSecret(t *testing.T) {
	type args struct {
		ctx   context.Context
		event SecretsmanagerTriggerPayload
		cfg   Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy path",
			args: args{
				ctx: context.TODO(),
				event: SecretsmanagerTriggerPayload{
					SecretARN: "arn:aws:secretsmanager:us-east-1:000000000000:secret:foo/bar-5BKPC8",
					Token:     "foo",
					Step:      "createSecret",
				},
				cfg: Config{
					SecretsmanagerClient: &mockSecretsmanagerClient{
						secretAWSCurrent: placeholderSecretUserStr,
					},
					DBClient:  clientDB{c: newMockSDKClient()},
					SecretObj: &SecretUser{},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := createSecret(tt.args.ctx, tt.args.event, tt.args.cfg); (err != nil) != tt.wantErr {
					t.Errorf("createSecret() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !tt.wantErr {
					secretInitial := placeholderSecretUser
					passwordInitial := secretInitial.Password
					secretInitial.Password = ""

					secretNew := getSecret(
						tt.args.cfg.SecretsmanagerClient.(*mockSecretsmanagerClient),
						"AWSPENDING",
						tt.args.event.Token,
					)
					passwordNew := secretNew.Password
					secretNew.Password = ""

					if passwordNew == passwordInitial || !reflect.DeepEqual(secretInitial, secretNew) {
						t.Errorf("generated secret does not match expectation")
					}
				}
			},
		)
	}
}

func Test_extractSecretObject1(t *testing.T) {
	type args struct {
		v      *secretsmanager.GetSecretValueOutput
		secret any
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantSecret *SecretUser
	}{
		{
			name: "happy path",
			args: args{
				v: &secretsmanager.GetSecretValueOutput{
					SecretString: &placeholderSecretUserStr,
				},
				secret: &SecretUser{},
			},
			wantErr:    false,
			wantSecret: &placeholderSecretUser,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if err := extractSecretObject(tt.args.v, tt.args.secret); (err != nil) != tt.wantErr {
					t.Errorf("extractSecretObject() error = %v, wantErr %v", err, tt.wantErr)
				}

				if !tt.wantErr {
					if !reflect.DeepEqual(tt.wantSecret, tt.args.secret) {
						t.Errorf("extractSecretObject() result does not match expectation")
					}
				}
			},
		)
	}
}