CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o watch-api main.go
ssh root@123.56.64.54 "cd /iWatch/ && sh stop.sh"

scp watch-api root@123.56.64.54:/iWatch/

ssh root@123.56.64.54 "cd /iWatch/ && sh run.sh"

go mod tidy