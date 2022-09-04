package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsZero(t *testing.T) {
	type args struct {
		data interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"int zero",
			args{
				data: 0,
			},
			true,
		},
		{
			"float zero",
			args{
				data: 0.0,
			},
			true,
		},
		{
			"string zero",
			args{
				data: "",
			},
			true,
		},
		{
			"string not zero",
			args{
				data: "0",
			},
			false,
		},
		{
			"array zero",
			args{
				data: [3]int{},
			},
			true,
		},
		{
			"map zero",
			args{
				data: make(map[string]interface{}),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsZero(tt.args.data); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsZeroForSlice(t *testing.T) {
	var s []int
	st := IsZero(s)
	assert.Equal(t, true, st)

	var s2 = make([]int, 0)
	st2 := IsZero(s2)
	assert.Equal(t, false, st2)
}

func TestStrArr2IntArr(t *testing.T) {
	type args struct {
		s          []string
		ignoreZero bool
	}
	tests := []struct {
		name    string
		args    args
		want    []int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "test1",
			args: args{
				s:          []string{"1", "2", "3", ""},
				ignoreZero: true,
			},
			want:    []int{1, 2, 3},
			wantErr: assert.NoError,
		},
		{
			name: "test2",
			args: args{
				s:          []string{"1", "b", "3", ""},
				ignoreZero: false,
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "test3",
			args: args{
				s:          []string{"1", "", "3", "2"},
				ignoreZero: false,
			},
			want:    []int{1, 0, 3, 2},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StrArr2IntArr(tt.args.s, tt.args.ignoreZero)
			if !tt.wantErr(t, err, fmt.Sprintf("StrArr2IntArr(%v, %v)", tt.args.s, tt.args.ignoreZero)) {
				return
			}
			assert.Equalf(t, tt.want, got, "StrArr2IntArr(%v, %v)", tt.args.s, tt.args.ignoreZero)
		})
	}
}

func TestWrapSqlLike(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				v: "HAHA",
			},
			want: "%HAHA%",
		},
		{
			name: "test2",
			args: args{
				v: "  HAHA",
			},
			want: "%HAHA%",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, WrapSqlLike(tt.args.v), "WrapSqlLike(%v)", tt.args.v)
		})
	}
}
