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


