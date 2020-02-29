package builder

import "testing"

func TestDarwin_Output(t *testing.T) {
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
			want: "test",
		},
		{
			name: "386",
			fields: fields{
				arch:   "386",
				output: "test",
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewDarwin(tt.fields.arch, tt.fields.output)
			if got := b.Output(); got != tt.want {
				t.Errorf("Darwin.Output() = %v, want %v", got, tt.want)
			}
		})
	}
}
