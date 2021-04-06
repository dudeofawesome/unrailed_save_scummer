package internal

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/gen2brain/beeep"
	"github.com/getlantern/systray"
)

var saveDir = ""
var backupsDir = ""
var assetDir = ""
var maxBackups = 5

func Setup() {
  log.Println("Starting")

  user, _ := user.Current()
  cwd, _ := os.Getwd()

  switch runtime.GOOS {
    case "darwin":
      saveDir = user.HomeDir + "/Library/Application Support/UnrailedGame/GameState/AllPlayers/SaveGames/"

      assetDir = path.Join(cwd, "Resources")
      if _, err := os.Stat(assetDir); os.IsNotExist(err) {
        assetDir = path.Join(cwd, "assets")
      }
    case "windows":
      saveDir = user.HomeDir + "\\AppData\\Local\\Daedalic Entertainment GmbH\\Unrailed\\GameState\\AllPlayers\\SaveGames\\"
        assetDir = path.Join(cwd, "assets")
    default:
      saveDir = user.HomeDir + "/.local/share/UnrailedGame/GameSate/AllPlayers/SaveGames/"
        assetDir = path.Join(cwd, "assets")
  }
  backupsDir = path.Join(saveDir, "backups")

  go setupFileWatcher()
  setupTrayIcon()
}

func setupTrayIcon() {
  log.Println("Setup tray icon")

  iconNixData, _ := os.ReadFile(path.Join(assetDir, "icon-template.png"))
  iconWinData, _ := os.ReadFile(path.Join(assetDir, "icon.ico"))

  systray.Run(func() {
    switch runtime.GOOS {
    case "windows":
      systray.SetIcon(iconWinData)
    case "darwin":
      systray.SetTemplateIcon(iconNixData, iconNixData)
    default:
      systray.SetIcon(iconNixData)
    }
    systray.SetTooltip("Unrailed Save Scummer")

    mAbout := systray.AddMenuItem("About Unrailed Save Scummer", "The fastest save scummer in the west")
    mQuit := systray.AddMenuItem("Quit", "Quit the program")
    go func() {
      select {
        case <-mAbout.ClickedCh:
          log.Println("About")
          // Open repo page in browser
          var cmd string
          var args []string
          switch runtime.GOOS {
            case "windows":
                cmd = "cmd"
                args = []string{"/c", "start"}
            case "darwin":
                cmd = "open"
            default:
                cmd = "xdg-open"
          }
          args = append(args, "https://github.com/dudeofawesome/unrailed-save-scummer")
          exec.Command(cmd, args...).Start()
        case <-mQuit.ClickedCh:
          log.Println("Quitting")
          systray.Quit()
      }
    }()
  }, func() {})
}

func setupFileWatcher() {
  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    log.Fatal(err)
  }
  defer watcher.Close()

  done := make(chan bool)
  go func() {
    defer handlePanic()

    for {
      select {
        case event, ok := <-watcher.Events:
          if !ok {
            return
          }

          _, basepath := path.Split(event.Name)

          _, err := os.Stat(event.Name);
          fileExists := os.IsExist(err)

          re := regexp.MustCompile(`SLOT(\d+)\.sav`)
          slotNumTest := re.FindAllStringSubmatch(basepath, -1)

          // make sure we're talking about a save game
          if (slotNumTest != nil) {
            slotNum, _ := strconv.Atoi(slotNumTest[0][1])

            if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
              // save file was created or modified
              log.Println("modified file:", event.Name)
              beeep.Notify("Unrailed Save Scummer", fmt.Sprintf("Backing up slot %d", slotNum), path.Join(assetDir, "/icon.png"))
              rotateSaves(slotNum)
              backupSave(slotNum)
            } else if
                event.Op&fsnotify.Remove == fsnotify.Remove ||
                // macOS seems to report deletion events as renames
                (runtime.GOOS == "darwin" && event.Op&fsnotify.Rename == fsnotify.Rename && !fileExists) {
              // save file was deleted (by the game)
              log.Println("deleted file:", event.Name)

              backupFile := path.Join(backupsDir, fmt.Sprintf("SLOT%d-0.sav", slotNum))
              log.Println(backupFile)
              _, err := os.Stat(backupFile)
              log.Println(err)

              if err == nil {
                restoreSave(slotNum, 0)
                beeep.Notify("Unrailed Save Scummer", fmt.Sprintf("Restored slot %d", slotNum), path.Join(assetDir, "/icon.png"))
              } else {
                beeep.Alert("Unrailed Save Scummer", fmt.Sprintf("No backup for slot %d to restore!", slotNum), path.Join(assetDir, "/icon.png"))
              }
            }
          }

        case err, ok := <-watcher.Errors:
          if !ok {
            return
          }
          log.Println("error:", err)
      }
    }
  }()

  err = watcher.Add(saveDir)
  if err != nil {
    log.Fatal(err)
  }
  <-done
}

func backupSave(saveSlot int)  {
  log.Println("Backing up save", saveSlot)

  // make sure backup dir exists
  if _, err := os.Stat(backupsDir); os.IsNotExist(err) {
    os.Mkdir(backupsDir, 0755)
  }

  // open read / write streams for save backup
  src, err := os.Open(path.Join(saveDir, fmt.Sprintf("SLOT%d.sav", saveSlot)))
  if err != nil { log.Panicf("Couldn't read SLOT%d.sav", saveSlot) }
  dest, err := os.Create(path.Join(backupsDir, fmt.Sprintf("SLOT%d-0.sav", saveSlot)))
  if err != nil { log.Panicf("Couldn't create backups/SLOT%d-0.sav", saveSlot) }
  io.Copy(dest, src)
  src.Close()
  dest.Close()
}

func rotateSaves(saveSlot int) {
  log.Println("Rotating backups for save", saveSlot)
  files, _ := os.ReadDir(backupsDir)

  // filter file array for valid save backups
  files = ArrayFilterDirEntry(
    files,
    func(file fs.DirEntry) bool {
      return file.Type().IsRegular() &&
        strings.HasPrefix(file.Name(), fmt.Sprintf("SLOT%d-", saveSlot)) &&
        strings.HasSuffix(file.Name(), ".sav")
    },
  )

  for i := len(files) - 1; i >= 0; i-- {
    file := files[i];
    re := regexp.MustCompile(`SLOT\d+-(\d+)\.sav`)
    backupNumber, _ := strconv.Atoi(re.FindAllStringSubmatch(file.Name(), -1)[0][1])

    // rotate save up one backup slot, or delete it if it's too old
    if (backupNumber < maxBackups - 1) {
      os.Rename(
        path.Join(backupsDir, file.Name()),
        path.Join(backupsDir, fmt.Sprintf("SLOT%d-%d.sav", saveSlot, backupNumber + 1)),
      )
    } else {
      os.Remove(path.Join(backupsDir, file.Name()))
    }
  }
}

func restoreSave(saveSlot int, backupSlot int) {
  log.Println("Restoring save", saveSlot)

  // open read / write streams for save backup
  src, err := os.Open(path.Join(backupsDir, fmt.Sprintf("SLOT%d-0.sav", saveSlot)))
  if err != nil { log.Panicf("Couldn't read backups/SLOT%d-0.sav", saveSlot) }
  dest, err := os.Create(path.Join(saveDir, fmt.Sprintf("SLOT%d.sav", saveSlot)))
  if err != nil { log.Panicf("Couldn't create SLOT%d.sav", saveSlot) }
  io.Copy(dest, src)
  src.Close()
  dest.Close()
}

func handlePanic() {
  if r := recover(); r != nil {
    beeep.Alert("Unrailed Save Scummer", fmt.Sprint(r), path.Join(assetDir, "/icon.png"))
  }
}
