//go:build windows

package archive

// oNoFollow is 0 on Windows: the platform has no O_NOFOLLOW open flag. The
// lexical containment check plus the refuseWriteThroughSymlink Lstat walk
// provide the cross-platform write-through-symlink defense.
const oNoFollow = 0
