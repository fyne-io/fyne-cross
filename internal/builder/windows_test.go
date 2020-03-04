package builder

import "testing"

func TestWindows_Output(t *testing.T) {
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
			want: "test.exe",
		},
		{
			name: "386",
			fields: fields{
				arch:   "386",
				output: "test",
			},
			want: "test.exe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewWindows(tt.fields.arch, tt.fields.output)
			if got := b.Output(); got != tt.want {
				t.Errorf("Windows.Output() = %v, want %v", got, tt.want)
			}
		})
	}
}
