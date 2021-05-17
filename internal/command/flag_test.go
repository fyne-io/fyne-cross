package command

import "testing"

func Test_envFlag_Set(t *testing.T) {
	tests := []struct {
		name    string
		value   []string
		wantLen int
		wantErr bool
	}{
		{
			name:    "simple env var",
			value:   []string{"CGO_ENABLED=1"},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "env var without value",
			value:   []string{"KEY="},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "env var with value containing =",
			value:   []string{"GOFLAGS=-mod=vendor"},
			wantLen: 1,
			wantErr: false,
		},
		{
			name:    "two env vars",
			value:   []string{"GOFLAGS=-mod=vendor", "KEY=value"},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "invalid",
			value:   []string{"GOFLAGS"},
			wantLen: 0,
			wantErr: true,
		},
		{
			name:    "env var with value containing comma",
			value:   []string{"GOFLAGS=https://goproxy.io,direct"},
			wantLen: 1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		ef := &envFlag{}
		t.Run(tt.name, func(t *testing.T) {
			for _, v := range tt.value {
				if err := ef.Set(v); (err != nil) != tt.wantErr {
					t.Errorf("envFlag.Set() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
			if len(*ef) != tt.wantLen {
				t.Errorf("envFlag len error = %v, wantLen %v", len(*ef), tt.wantLen)
			}
		})
	}
}
