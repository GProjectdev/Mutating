# 1단계: 빌드 스테이지 (Go 1.24 기반)
FROM golang:1.24 as builder

WORKDIR /app
COPY go.mod ./
COPY cmd/ ./cmd/

# 의존성 설치
RUN go mod tidy

# 정적 바이너리 빌드 (CGO 비활성화)
RUN CGO_ENABLED=0 GOOS=linux go build -o webhook-server ./cmd/

# 2단계: 실행 전용 스테이지 (경량 이미지)
FROM alpine:latest

# TLS 인증서 검증용 ca-certificates 추가
RUN apk add --no-cache ca-certificates

# 실행 디렉토리 설정
WORKDIR /root/

# 빌드 결과물만 복사
COPY --from=builder /app/webhook-server .

# 실행
CMD ["./webhook-server"]
