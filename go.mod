module github.com/amitsaha/gitbackup

go 1.17

require (
	github.com/99designs/keyring v1.1.6
	github.com/cli/oauth v0.8.0
	github.com/google/go-github/v34 v34.0.0
	github.com/ktrysmt/go-bitbucket v0.9.1
	github.com/migueleliasweb/go-github-mock v0.0.5
	github.com/mitchellh/go-homedir v1.1.0
	github.com/spf13/afero v1.2.2
	github.com/xanzy/go-gitlab v0.16.1
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/text v0.3.5 // indirect
)

require (
	github.com/danieljoos/wincred v1.0.2 // indirect
	github.com/dvsekhvalnov/jose2go v0.0.0-20200901110807-248326c1351b // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/google/go-github/v37 v37.0.0 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/k0kubun/pp v2.3.0+incompatible // indirect
	github.com/keybase/go-keychain v0.0.0-20190712205309-48d3d31d256d // indirect
	github.com/mattn/go-colorable v0.0.9 // indirect
	github.com/mattn/go-isatty v0.0.3 // indirect
	github.com/mitchellh/mapstructure v0.0.0-20180220230111-00c29f56e238 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	golang.org/x/crypto v0.0.0-20190701094942-4def268fd1a4 // indirect
	golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3 // indirect
	golang.org/x/sys v0.0.0-20190712062909-fae7ac547cb7 // indirect
	google.golang.org/appengine v1.4.0 // indirect
)

// https://github.com/99designs/keyring/pull/101/files
replace golang.org/x/sys v0.0.0-20190712062909-fae7ac547cb7 => golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e
