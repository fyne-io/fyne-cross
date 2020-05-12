package command

import (
	"reflect"
	"testing"
)

func Test_makeDefaultContext(t *testing.T) {
	type args struct {
		flags *CommonFlags
	}
	tests := []struct {
		name        string
		args        args
		wantEnv     []string
		wantLdFlags []string
	}{
		{
			name: "default",
			args: args{
				flags: &CommonFlags{
					Env: &envFlag{},
				},
			},
			wantEnv:     []string{},
			wantLdFlags: nil,
		},
		{
			name: "custom env",
			args: args{
				flags: &CommonFlags{
					Env: &envFlag{"TEST=true"},
				},
			},
			wantEnv:     []string{"TEST=true"},
			wantLdFlags: nil,
		},
		{
			name: "custom ldflags",
			args: args{
				flags: &CommonFlags{
					Env:     &envFlag{},
					Ldflags: "-X main.version=1.2.3",
				},
			},
			wantEnv:     []string{},
			wantLdFlags: []string{"-X main.version=1.2.3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := makeDefaultContext(tt.args.flags)
			if err != nil {
				t.Errorf("couldn't create Windows context: %v", err)
			}

			if !reflect.DeepEqual(ctx.Env, tt.wantEnv) {
				t.Errorf("expected env %s but got %s", tt.wantEnv, ctx.Env)
			}

			if !reflect.DeepEqual(ctx.LdFlags, tt.wantLdFlags) {
				t.Errorf("expected ldflags %s but got %s", tt.wantLdFlags, ctx.LdFlags)
			}
		})
	}
}
