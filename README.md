# mono-buy
This project is Monolithic Architecture project for trade in binance.

## 개발 시 실행 방법

```bash
go run .
```

## build 하는 법

```bash
go build
```

## Linux 용 build 하는 법

```bash
# 명령 프롬프트(cmd)에서:
GOOS=linux GOARCH=amd64 go build -o mono-buy

# PowerShell에서:
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o mono-buy
```

## Linux에서 백그라운드 실행하는 법

```bash
nohup ./mono-buy &
```

```bash
nohup ./mono-buy > /dev/null 2>&1 &
```

## 정상 종료 하는 방법

```bash
ps aux | grep mono-buy

# 정상 종료
kill -15 [PID번호]

# 강제 종료 (필요시)
kill -9 [PID번호]
```