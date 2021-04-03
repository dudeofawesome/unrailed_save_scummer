package internal

import "io/fs"

func ArrayFilterDirEntry(ss []fs.DirEntry, test func(fs.DirEntry) bool) (ret []fs.DirEntry) {
  for _, s := range ss {
    if test(s) {
      ret = append(ret, s)
    }
  }
  return
}
