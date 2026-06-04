//go:build unix

package archive

import "syscall"

// oNoFollow is OR'd into the open flags for regular-file extraction writes so
// the kernel refuses to follow a final-component symlink (CWE-59 write-through
// escape). On Unix it maps to syscall.O_NOFOLLOW; on Windows it is 0 because
// the platform has no equivalent open flag and the lexical + Lstat-walk
// defense in refuseWriteThroughSymlink carries the containment guarantee.
const oNoFollow = syscall.O_NOFOLLOW
