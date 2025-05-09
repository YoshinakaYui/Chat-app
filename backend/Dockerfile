FROM golang:1.24

# 作業ディレクトリを指定
WORKDIR /app

# 必要なパッケージをインストール
#RUN apk add --no-cache git

# AirとOgenのインストール
RUN go install github.com/ogen-go/ogen/cmd/ogen@latest
RUN go install github.com/air-verse/air@latest

# RUN ogen --version

ENV PATH="/go/bin:${PATH}"

# モジュールファイルをコピー
COPY go.mod go.sum ./
RUN go mod download

# API仕様書をコピーして、コードを自動生成
COPY api.yaml .
#RUN ogen generate -o ./gen --package api api.yaml

#RUN /go/bin/ogen generate -o ./gen --package api /app/api.yaml
RUN which ogen && ogen --version

# ソースコードをコピー
COPY . .

# アプリケーションのビルド
#RUN go build -o main .

# ポートを公開
EXPOSE 8080

# ホットリロード用のAirを使って実行
CMD ["air"]




# # ベースイメージとしてGoを使用
# FROM golang:1.20-alpine AS builder

# # 必要なパッケージをインストール
# RUN apk update && apk add --no-cache git curl build-base

# # 作業ディレクトリを設定
# WORKDIR /app

# # 環境変数の設定
# ENV GOPATH=/go
# ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH
# ENV GO111MODULE=on

# # Ogenをインストール
# RUN go install github.com/ogen-go/ogen/cmd/ogen@latest

# # モジュールファイルをコピーして依存関係を解決
# COPY go.mod go.sum ./
# RUN go mod download

# # アプリケーションソースコードをコピー
# COPY . .

# # Ogenが正しくインストールされているか確認（デバッグ用）
# RUN which ogen && ogen --version

# # Ogenでコード生成
# RUN ogen generate -o ./gen --package api api.yaml

# # アプリケーションビルド
# RUN go build -o main .

# # 実行環境として軽量なAlpine Linuxを使用
# FROM alpine:latest
# WORKDIR /app

# # バイナリをコピー
# COPY --from=builder /app/main /app/main

# # ポートを公開
# EXPOSE 8080

# # サーバーを実行
# CMD ["/app/main"]

