
[![GoDoc](https://godoc.org/github.com/ricardobranco777/go-userfaultfd?status.svg)](https://godoc.org/github.com/ricardobranco777/go-userfaultfd)
![Build Status](https://github.com/ricardobranco777/go-userfaultfd/actions/workflows/ci.yml/badge.svg)

Golang interface to Linux's [userfaultfd](https://man7.org/linux/man-pages/man2/userfaultfd.2.html) system call.

NOTES
- Must set `vm.unprivileged_userfaultfd` as user for some features.

Similar projects:
- https://github.com/loopholelabs/userfaultfd-go

More information at:
- https://github.com/torvalds/linux/blob/master/mm/userfaultfd.c
- https://docs.kernel.org/admin-guide/mm/userfaultfd.html
- https://man.archlinux.org/man/userfaultfd.2.en
- https://man.archlinux.org/man/ioctl_userfaultfd.2.en
- https://www.cons.org/cracauer/cracauer-userfaultfd.html
- https://lwn.net/Articles/897260/
