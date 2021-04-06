package main

import (
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/dudeofawesome/unrailed_save_scummer/internal"
)

func main()  {
  // log, err := os.Create("/Users/dudeofawesome/unrailed_save_scummer.log")
  // if err != nil {
  //   log.WriteString(fmt.Sprint(err) + "\n")
  // }
  // log.WriteString("LOGSSSS\n")
  // log.WriteString(fmt.Sprint(os.Args) + "\n")
  // log.WriteString(fmt.Sprint(app_dir) + "\n")
  // log.WriteString(fmt.Sprint(os.Getwd()) + "\n")
  // log.Close()

  if runtime.GOOS == "darwin" {
    wd := strings.Split(os.Args[0], "/")
    app_dir := strings.Join(wd[0:len(wd) - 3], "/")
    os.Chdir(path.Join(app_dir, "Contents"))
  }

  internal.Setup()
}
