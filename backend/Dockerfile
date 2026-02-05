# 只複製預先編譯好的二進位檔
# 原因：本機先執行 build.sh 編譯，Docker 不需要安裝任何依賴
FROM alpine:latest

WORKDIR /app
COPY accountbook-server .
RUN chmod +x accountbook-server

# 建立資料目錄
RUN mkdir -p /app/data

EXPOSE 8080
CMD ["./accountbook-server"]
