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
