package builder

import "testing"

func TestLinux_Output(t *testing.T) {
	type fields struct {
		os     string
		arch   string
		output string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "amd64",
			fields: fields{
				arch:   "amd64",
				output: "test",
			},
			want: "test-linux-amd64",
		},
		{
			name: "386",
			fields: fields{
				arch:   "386",
				output: "test",
			},
			want: "test-linux-386",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewLinux(tt.fields.arch, tt.fields.output)
			if got := b.Output(); got != tt.want {
				t.Errorf("Linux.Output() = %v, want %v", got, tt.want)
			}
		})
	}
}
