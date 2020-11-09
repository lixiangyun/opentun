set GOOS=linux
set GOARCH=amd64

go build -o tun

set GOOS=linux
set GOARCH=arm64

go build -o tun_arm64

scp tun root@122.9.82.162:/root
scp tun root@192.168.3.30:/root
scp tun_arm64 root@192.168.3.27:/root
