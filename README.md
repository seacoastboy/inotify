inotify
=======

A fixed exp/inotify package for go.

Changes made:
 * Clean up watch info if IN_IGNORED is received
 * Make RemoveWatch safe for concurrent access
 * Make Close() reliably (max 500 ms delay before the syscall times out)

Documentation at http://go.pkgdoc.org/github.com/mb0/inotify
