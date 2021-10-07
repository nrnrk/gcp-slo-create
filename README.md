## Quick Start

```shell
go install github.com/nrnrk/gcp-slo-create@latest

# トークンの取得
GCP_SLO_SETTER_TOKEN=`gcloud auth print-access-token`

# 対象となるプロジェクト・ロードバランシングを指定
PROJECT_ID=<your-project-id>
URL_MAP_NAME=<your-load-balancing-name>

gcp-slo-create -project-id ${PROJECT_ID} -url-map-name ${URL_MAP_NAME} -token ${GCP_SLO_SETTER_TOKEN}
```

## 各引数について

### project id

以下のコマンドで表示されるリストの `PROJECT_ID` のフィールドから1つ選択してください。

```shell
gcloud projects list
```

### url map name

以下のコマンドで表示されるリストの `NAME` のフィールドから1つ選択してください。

```shell
gcloud compute url-maps list
```