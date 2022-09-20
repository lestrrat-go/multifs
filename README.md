multifs
=======

Create an fs.FS instance that "mounts" other fs.FS.

Note: extremely alpha quality.

# SYNOPSIS

```go
fs1 := ....
fs2 := ....

var fs multifs.FS

fs.Mount("/fs1", fs1)
fs.Mount("/fs2", fs2)

f, err := fs.Open("/fs1/foo/bar/baz.txt")
```
