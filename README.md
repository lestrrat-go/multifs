multifs
=======

Create an fs.FS instance that "mounts" other fs.FS.

As of this writing this module is really aimed at one-off tools and testing,
and thus does not really prioritize on efficiency

Note: extremely alpha quality.

# SYNOPSIS

```go
fs1 := ....
fs2 := ....

var fs multifs.FS

fs.Mount("/fs1", fs1)
fs.Mount("/fs2", fs2)

f1, err := fs.Open("/fs1/foo/bar/baz.txt")
f2, err := fs.Open("/fs2/quux/corge/grault.txt")
```

# Path Expansion

Given a prefix of "/prefix", call to `(multifs.FS).Open("/prefix/file.txt")` results in the
backend filesystem receiving "file.txt" as the argument to its `Open()` method.

This means that you should most likely not just use `os.Open` in the backend.
Instead you should most likely use `os.DirFS` which takes care of the prefix.

Let's say you have a file name `"/root/temp/12345/file.txt"`. and you want this to be availble
in the `multifs.FS` as `"/prefix/file.txt"`. Then you should do the follwing:

```go
fileFS, _ := os.DirFS("/root/temp/12345")

var fs multifs.FS

fs.Mount("/prefix", fileFS)
```

The use of `os.DirFS` makes sure that it takes care of replacing the prefix
with the actual file system prefix for you.
