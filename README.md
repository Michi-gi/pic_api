# pic_api
これは[pic_fromt](https://github.com/Michi-gi/pic_fromt)から呼び出されること前提の、画像サイトアクセス用のAPIです。このAPIで下記が解決されます。
- バラバラな各画像サイトの呼び出し方を統一することができる。
- 自サイトからの呼び出しでCORSを回避することができる(クライアントと同一Originにできる)。

## 機能
このAPIは次の機能があります。

|endpoint|機能|
|---|---|
|/picprofile|画像情報取得|
|/authorprofile|著者情報取得|
|/picbyauthor|著者の画像一覧|
|/judgesite|画像サイト判別|
|/download|画像ダウンロード|

## 利用方法
コードをビルドして作成された実行ファイルを実行してください。またDockerfileも用意しています。

## 対応サイト
現時点で対応しているサイトは次の通り

- [Pixiv](https://www.pixiv.net/) ([pic_api_pixiv](https://github.com/Michi-gi/pic_api_pixiv)が必要)

## 利用ライブラリー
APIのリクエストのルーティングに[chi](https://github.com/go-chi/chi)を利用しています。