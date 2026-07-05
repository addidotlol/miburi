# miburi

mouse gestures for gnome using evdev.

hold the forward button on your mouse and:

* move up: open the activities overview
* move left or right: switch workspaces

## how it works

miburi grabs your mouse and mirrors it through a virtual device. when you
hold down the forward button on your mouse, it initiates the gesture detection.
once a mouse movement is detected it will send key presses (super, super+page up,
super+page down) from a virtual keyboard.

if no movement is detected, it will send a normal forward input.

## build

    task build

needs [go](https://go.dev) and [go-task](https://taskfile.dev). run with sudo, it needs /dev/input and /dev/uinput.

## packages

    task rpm
    task deb

builds an rpm (rpmbuild) or deb (dpkg-deb) into dist/. install one, then:

    sudo systemctl enable --now miburi

## flags

* `-list` list input devices
* `-device` input device path, auto-detected when empty
* `-button` trigger button, default BTN_EXTRA
* `-threshold` movement needed to fire a gesture, default 250
* `-overview` keys for the up gesture, default super
* `-left` keys for the left gesture, default super+pageup
* `-right` keys for the right gesture, default super+pagedown

combos are plus-separated key names, like `ctrl+super+up`. they have to
match shortcuts that are actually bound in gnome, check with:

    gsettings list-recursively org.gnome.desktop.wm.keybindings | grep workspace

the service reads flags from MIBURI_OPTS in /etc/sysconfig/miburi.

## license

gpl-3.0
