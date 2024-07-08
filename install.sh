#!/bin/env bash
#

[ -z $(command -v go) ] && {
  echo "[X] go binary could not be found"
  exit 1
}

[ -n "$GOPATH" ] && BINPATH="$GOPATH/bin"
[ -z "$BINPATH" ] && {
  echo "[!] \$GOPATH could not be found. falling back to \$HOME/go/bin."
  BINPATH="$HOME/go/bin"
}
mkdir -p $BINPATH

echo "* Building binary into $BINPATH..."
go build -o "$BINPATH/lowdb" ./cmd

[ -n "$SUDO_USER" ] && echo "PASHM: $SUDO_USER" #USER=$SUDO_USER

echo "* Writing systemd unit file for user $USER into /tmp/lowdb.service"
cat <<EOF > /tmp/lowdb.service
[Unit]
Description=LowDB systemd service.

[Service]
ExecStart=$BINPATH/lowdb serve
User=$USER

[Install]
WantedBy=multi-user.target
EOF


if [ "$EUID" -ne 0 ]; then
  echo "[!] Not running with root access. please run these commands:"
  echo "sudo cp /tmp/lowdb.service /etc/systemd/system/lowdb.service"
  echo "sudo systemctl daemon-reload"
  echo "sudo systemctl start lowdb.service"
else
  cp /tmp/lowdb.service /etc/systemd/system/lowdb.service
  systemctl daemon-reload
  systemctl start lowdb.service
fi

echo "* Done"
