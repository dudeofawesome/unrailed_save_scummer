#!/usr/bin/env bash

mac_dir="./bin/macOS"
mac_app_dir="$mac_dir/Unrailed Save Scummer.app"
mac_app_contents_dir="$mac_app_dir/Contents"

rm -rf "./bin"

pkger

mkdir -p "$mac_app_contents_dir/"{MacOS,Resources}

if [[ "$OSTYPE" == "darwin"* ]]; then
  # build ICNS
  mkdir "./assets/images/icon.iconset"
  for size in 128 256; do
    convert "./assets/images/icon.png" -resize "${size}x${size}" "./assets/images/icon.iconset/icon_${size}x${size}.png"
  done
  png2icns "./assets/images/icon.icns" "./assets/images/icon.iconset/"*.png
  cp "./assets/images/icon.icns" "$mac_dir/.VolumeIcon.icns"
  cp "./assets/images/icon.icns" "$mac_app_contents_dir/Resources/icon.icns"

  # package images
  cp "./assets/images/icon.png" "$mac_app_contents_dir/Resources/"

  # template Info.plist
  eval "cat <<EOF
  $(<"./assets/Info.plist")
  EOF
  " 2> /dev/null > "$mac_app_contents_dir/Info.plist"

  # # Prepare DMG assets
  # cp "./assets/dmg-contents/USS.DS_Store" "$mac_dir/.DS_Store"
  # ln -s "/Applications" "$mac_dir/Applications"
fi

# build
out="./bin/unrailed-save-scummer"
case "$OSTYPE" in
linux)
  out="./bin/unrailed-save-scummer-amd64-linux" ;;
darwin*)
  out="$mac_app_contents_dir/MacOS/unrailed-save-scummer-amd64-darwin" ;;
*)
  echo "OS is $OSTYPE"
  out="./bin/unrailed-save-scummer-amd64-windows.exe" ;;
esac

echo "Building binary"
go build -o "$out" .

case "$OSTYPE" in
darwin*)
  # Package macOS app into DMG
  width=600
  height=400
  create-dmg \
    --volname "Unrailed Save Scummer" \
    --volicon "./assets/images/icon.icns" \
    --window-size $width $height \
    --icon-size $(expr $height / 4) \
    --icon "Unrailed Save Scummer.app" $(expr $width / 4 - $height / 8) $(expr $height / 2 - $height / 6) \
    --hide-extension "Unrailed Save Scummer.app" \
    --app-drop-link $(expr $width \* 3 / 4 - $height / 8) $(expr $height / 2 - $height / 6) \
    --hdiutil-quiet \
    --no-internet-enable \
    "./bin/unrailed-save-scummer-amd64-mac.dmg" "$mac_app_dir"
  ;;
linux)
  ;;
*)
  ;;
esac
