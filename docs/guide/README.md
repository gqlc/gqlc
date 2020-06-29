# Installing

The recommended way of installing `gqlc` is by downloading one of the
[pre-built binaries](https://github.com/gqlc/gqlc/releases).

For the more DIY inclined, however, `gqlc` can also be installed from source by
following the below steps.

## Building Source

Some prerequisites for building and installing `gqlc` from source are:

* [Go](https://golang.org)
* [git](https://git-scm.com)

After installing the prereqs, you need to clone the source code:
```bash
git clone git@github.com:gqlc/gqlc.git # or https://github.com/gqlc/gqlc.git
cd gqlc

# If you want to install a specific version
# Then run the following
git checkout vTHE.SPECIFIC.VERSION
```

Finally, installing is as easy as running the following command:
```bash
go install

# If installing a specific version from the previous step
# Then you need to use the following
go install -ldflags="-X github.com/gqlc/gqlc/cmd.version=THE.SPECIFIC.VERSION"
```
