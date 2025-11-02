
Golang interface to Linux's [userfaultfd](https://man7.org/linux/man-pages/man2/userfaultfd.2.html) system call.

TODO
- Add ReadMsg to read events.
- Implement higher level API.
- Support /dev/userfaultfd

NOTES
- Must set `vm.unprivileged_userfaultfd` as user for some features.

Tested on:
| Arch | notes |
|---|---|
| arm64 | Debian 13 with kernel 6.12 and `CONFIG_USERFAULTFD` not set |
| amd64 | Fedora 43 with kernel 6.17 |
| ppc64le | SLES 15-SP3 with kernel 5.3 |
| s390x | SLES 16.0 with kernel 6.12 |

Similar projects:
- https://github.com/loopholelabs/userfaultfd-go

More information at:
- https://docs.kernel.org/admin-guide/mm/userfaultfd.html
- https://www.cons.org/cracauer/cracauer-userfaultfd.html
