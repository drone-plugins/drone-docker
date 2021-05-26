package docker

import (
	"reflect"
	"testing"
)

func Test_commandPush(t *testing.T) {
	type args struct {
		build Build
		tag   string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "removes everything after last colon in CacheFrom[0]",
			args: args{
				build: Build{
					CacheBuilder: true,
					CacheFrom:    []string{"harbor.shipttech.com/buildcache/kubedashian-api:ncooke_ch339348_implement-caching-for-docker-in-docker-builds"},
				},
				tag: "foo",
			},
			want: "/usr/local/bin/docker push harbor.shipttech.com/buildcache/kubedashian-api:foo",
		},
		{
			name: "returns exact value plus tag in CacheRepo",
			args: args{
				build: Build{
					CacheBuilder: true,
					CacheRepo:    "harbor.shipttech.com/buildcache/kubedashian-api",
				},
				tag: "foo",
			},
			want: "/usr/local/bin/docker push harbor.shipttech.com/buildcache/kubedashian-api:foo",
		},
		{
			name: "handles no colon in CacheFrom[0]",
			args: args{
				build: Build{
					CacheBuilder: true,
					CacheFrom:    []string{"harbor.shipttech.com/buildcache/kubedashian-api"},
				},
				tag: "foo",
			},
			want: "/usr/local/bin/docker push harbor.shipttech.com/buildcache/kubedashian-api:foo",
		},
		{
			name: "returns Repo if CacheBuilder false",
			args: args{
				build: Build{
					CacheBuilder: false,
					CacheFrom:    []string{"no"},
					CacheRepo:    "not this",
					Repo:         "harbor.shipttech.com/buildcache/kubedashian-api",
				},
				tag: "foo",
			},
			want: "/usr/local/bin/docker push harbor.shipttech.com/buildcache/kubedashian-api:foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := commandPush(tt.args.build, tt.args.tag); !reflect.DeepEqual(got.String(), tt.want) {
				t.Errorf("\ngot  = %v\nwant = %v", got.String(), tt.want)
			}
		})
	}
}
