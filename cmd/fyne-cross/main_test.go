package main

import (
	"reflect"
	"testing"
)

func Test_parseTargets(t *testing.T) {
	type args struct {
		targetList string
	}
	tests := []struct {
		name    string
		args    args
		want    [][2]string
		wantErr bool
	}{
		{
			name:    "Target list cannot be empty",
			args:    args{targetList: ""},
			want:    [][2]string{},
			wantErr: true,
		},
		{
			name:    "Invalid target",
			args:    args{targetList: "invalid/amd64"},
			want:    [][2]string{},
			wantErr: true,
		},
		{
			name:    "Invalid target 2",
			args:    args{targetList: "linux/amd64,invalid/amd64"},
			want:    [][2]string{{"linux", "amd64"}},
			wantErr: true,
		},
		{
			name:    "Invalid target 3",
			args:    args{targetList: "linux/*amd64"},
			want:    [][2]string{},
			wantErr: true,
		},
		{
			name:    "Valid target",
			args:    args{targetList: "linux/amd64"},
			want:    [][2]string{{"linux", "amd64"}},
			wantErr: false,
		},
		{
			name:    "Valid targets trim space",
			args:    args{targetList: "linux/amd64, darwin/386"},
			want:    [][2]string{{"linux", "amd64"}, {"darwin", "386"}},
			wantErr: false,
		},
		{
			name:    "Valid wildcard targets",
			args:    args{targetList: "linux/*"},
			want:    [][2]string{{"linux", "amd64"}, {"linux", "386"}, {"linux", "arm"}, {"linux", "arm64"}},
			wantErr: false,
		},
		{
			name:    "Android",
			args:    args{targetList: "android"},
			want:    [][2]string{{"android", ""}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTargets(tt.args.targetList)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTargets() error = %#+v, wantErr %#+v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTargets() = %#+v, want %#+v", got, tt.want)
			}
		})
	}
}
