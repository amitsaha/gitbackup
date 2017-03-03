[![Build Status](https://travis-ci.org/amitsaha/gitbackup.svg?branch=master)](https://travis-ci.org/amitsaha/gitbackup)

## Building

Install `gb` using the instructions [here](https://getgb.io/) and then also install
the [vendor](https://godoc.org/github.com/constabulary/gb/cmd/gb-vendor) plugin using:
`go install github.com/constabulary/gb/cmd/gb-vendor`.

Next, we will fetch the package using `gb vendor` as:

```
gb vendor fetch github.com/mitchellh/go-homedir github.com/google/go-github/github
```
This will create a `vendor` sub-directory which will have the dependencies (including their dependencies).

Then, we can build it:

```
gb build all
```

