FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Ho_Chi_Minh

COPY go-backend-linux ./go-backend
RUN chmod +x ./go-backend
COPY configs/configs.yaml ./configs/
COPY Templates ./Templates

EXPOSE 44318

CMD ["./go-backend"]
