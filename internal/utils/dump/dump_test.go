package dump

import "testing"

func Test_title(t *testing.T) {
	type args struct {
		title  string
		length int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1", args{"1234", 10}, "-- 1234 --"},
		{"2", args{"12345", 10}, "-- 12345 -"},
		{"3", args{"123456", 10}, "- 123456 -"},
		{"4", args{"1234567", 10}, "1234567"},
		{"5", args{"123456789", 10}, "123456789"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := title(tt.args.title, tt.args.length); got != tt.want {
				t.Errorf("title() = %v, want %v", got, tt.want)
			}
		})
	}
}
