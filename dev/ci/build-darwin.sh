#!/usr/bin/env bash

app_dir='./bin/Unrailed Save Scummer.app/Contents'
rm -rf "$app_dir"

pkger

mkdir -p "$app_dir"
mkdir "$app_dir/MacOS"
mkdir "$app_dir/Resources"

# build ICNS
mkdir "./assets/images/icon.iconset"
for size in 128 256 512; do
  sips -z "$size" "$size" "./assets/images/icon.png" --out "./assets/images/icon.iconset/icon_${size}x${size}.png"
done
png2icns "$app_dir/Resources/icon.icns" "./assets/images/icon.iconset/"*.png

# package images
cp "./assets/images/icon.png" "$app_dir/Resources/"

# template Info.plist
eval "cat <<EOF
$(<"./assets/Info.plist")
EOF
" 2> /dev/null > "$app_dir/Info.plist"

# build
GOOS=darwin GOARCH=amd64 go build -o "$app_dir/MacOS/app-amd64-darwin" .
